package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Constants for archetypes
const (
	ArchetypeSummarization = "Summarization"
	ArchetypeQA            = "QuestionAnswering"
	ArchetypeUnknown       = "Unknown"
)

var (
	goal         string
	returnFormat string
	warnings     string
	contextDump  string
	confirm      bool
)

// detectArchetype analyzes the goal to determine the likely user intent.
func detectArchetype(goal string) string {
	lowerGoal := strings.ToLower(goal)

	// Keywords for Question Answering (prioritized)
	qaKeywords := []string{"what is", "explain", "how does", "list", "compare", "does it", "can i", "where", "who", "when", "why"}
	for _, keyword := range qaKeywords {
		if strings.Contains(lowerGoal, keyword) {
			return ArchetypeQA
		}
	}

	// Keywords for Summarization
	summarizeKeywords := []string{"summarize", "summary", "overview", "tldr", "key points", "abstract", "give me the gist"}
	for _, keyword := range summarizeKeywords {
		if strings.Contains(lowerGoal, keyword) {
			return ArchetypeSummarization
		}
	}

	return ArchetypeUnknown
}

// getGuidance returns helpful tips based on the detected archetype.
func getGuidance(archetype string) []string {
	switch archetype {
	case ArchetypeSummarization:
		return []string{
			"Tip: Consider specifying desired length (e.g., 'one paragraph', 'bullet points').",
			"Tip: Mention the target audience if applicable.",
			"Tip: Focus on specific aspects if needed (e.g., 'summarize security controls').",
		}
	case ArchetypeQA:
		return []string{
			"Tip: Ensure your question is specific for better answers.",
			"Tip: Use terminology likely found in the provided context.",
			"Tip: If asking about multiple things, consider separate prompts.",
		}
	default:
		return nil // No guidance for unknown archetype
	}
}

// getEnrichmentSuggestions provides refined suggestions for goal and return format.
func getEnrichmentSuggestions(archetype string, originalGoal string, originalFormat string) (suggestedGoal string, suggestedFormats []huh.Option[string]) {
	switch archetype {
	case ArchetypeSummarization:
		suggestedGoal = "Summarize the key requirements and obligations mentioned in the provided documents."
		suggestedFormats = []huh.Option[string]{
			huh.NewOption("Bulleted list of key points", "Bulleted list of key points"),
			huh.NewOption("Concise paragraph overview", "Concise paragraph overview"),
		}
		// Add original format as an option if it's different and not empty
		if originalFormat != "" && originalFormat != "Bulleted list of key points" && originalFormat != "Concise paragraph overview" {
			suggestedFormats = append(suggestedFormats, huh.NewOption("Keep: "+originalFormat, originalFormat))
		} else if originalFormat == "" {
			// Add a generic keep option if original was empty
			suggestedFormats = append(suggestedFormats, huh.NewOption("Keep original (empty)", ""))
		}

	case ArchetypeQA:
		// Prepend standard framing to the user's original question/goal
		suggestedGoal = fmt.Sprintf("Based *only* on the provided documents, answer the question: %s", originalGoal)
		suggestedFormats = []huh.Option[string]{
			huh.NewOption("Direct answer", "Direct answer"),
			huh.NewOption("Answer with citations to relevant sections", "Answer with citations"),
			huh.NewOption("Extract relevant quotes supporting the answer", "Extract relevant quotes"),
		}
		// Add original format as an option if it's different and not empty
		if originalFormat != "" && originalFormat != "Direct answer" && originalFormat != "Answer with citations" && originalFormat != "Extract relevant quotes" {
			suggestedFormats = append(suggestedFormats, huh.NewOption("Keep: "+originalFormat, originalFormat))
		} else if originalFormat == "" {
			suggestedFormats = append(suggestedFormats, huh.NewOption("Keep original (empty)", ""))
		}

	default:
		// No suggestions for unknown archetype
		return originalGoal, nil
	}

	// Ensure the original format is the default selection if it exists in the options
	// If not, the first suggestion becomes the default. Huh handles default selection.

	return suggestedGoal, suggestedFormats
}

func main() {

	// --- Initial Form ---
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Step 1: Initial Prompt Details"))
	initialForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Goal").
				Placeholder("e.g., Summarize the key requirements").
				CharLimit(500).
				Value(&goal),
			huh.NewText().
				Title("Return Format").
				Placeholder("e.g., Bulleted list").
				CharLimit(500).
				Value(&returnFormat),
			huh.NewText().
				Title("Warnings").
				Placeholder("e.g., Exclude information about XYZ").
				CharLimit(500).
				Value(&warnings),
			huh.NewText().
				Title("Context Dump").
				Placeholder("e.g., Paste relevant sections of compliance docs here").
				CharLimit(500).
				Value(&contextDump),
		),
	)

	err := initialForm.Run()
	if err != nil {
		// Check for CTRL+C
		if err == huh.ErrUserAborted {
			fmt.Println("\nOperation cancelled by user.")
			os.Exit(0)
		}
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
	// Store original values before potential enrichment
	originalGoal := goal
	originalFormat := returnFormat

	// --- Archetype Detection and Guidance ---
	detectedArchetype := detectArchetype(goal)
	if detectedArchetype != ArchetypeUnknown {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("\nPrompt Guidance:")) // Green color for header
		guidance := getGuidance(detectedArchetype)
		for _, tip := range guidance {
			fmt.Println(lipgloss.NewStyle().Faint(true).Render("- " + tip))
		}
		fmt.Println() // Add a blank line after guidance

		// --- Enrichment Step ---
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("Step 2: Refine Prompt (Optional)"))
		suggestedGoal, suggestedFormats := getEnrichmentSuggestions(detectedArchetype, originalGoal, originalFormat)

		// Use the suggested goal as the new value for the text field
		goal = suggestedGoal
		// The selected format will update the main returnFormat variable directly

		enrichmentForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Detected Intent: "+detectedArchetype).Description("We detected a potential intent. You can refine the Goal and Return Format below."),
				huh.NewText().
					Title("Refined Goal").
					Description("Suggested goal based on detection. Edit as needed.").
					Value(&goal), // Bind directly to the main goal variable
				huh.NewSelect[string](). // Use type parameter [string]
								Title("Suggested Return Format").
								Options(suggestedFormats...). // Use the generated options
								Value(&returnFormat),         // Bind directly to the main returnFormat variable
			),
		)

		err = enrichmentForm.Run()
		if err != nil {
			// Check for CTRL+C
			if err == huh.ErrUserAborted {
				fmt.Println("\nOperation cancelled by user.")
				os.Exit(0)
			}
			// If error occurs here, maybe revert to original goal/format? Or just exit.
			fmt.Println("Error during enrichment:", err)
			os.Exit(1)
		}
	} else {
		// If archetype is unknown, just print a separator
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("\n---"))
	}

	// --- Confirmation Step ---
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("\nStep 3: Confirm Generation"))
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate prompt with current details?").
				Affirmative("Yes!").
				Negative("No!").
				Value(&confirm),
		),
	)
	err = confirmForm.Run()
	if err != nil || !confirm {
		if err == huh.ErrUserAborted || !confirm {
			fmt.Println("Prompt generation cancelled.")
			os.Exit(0)
		}
		fmt.Println("Confirmation error:", err)
		os.Exit(1)
	}

	// --- Spinner and Final Output ---
	preparePrompt := func() {
		time.Sleep(1 * time.Second)
	}

	_ = spinner.New().Title("Preparing your prompt...").Action(preparePrompt).Run()

	{
		var sb strings.Builder
		labelStyle := lipgloss.NewStyle().Bold(true)
		contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

		fmt.Fprintf(&sb,
			"%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n",
			labelStyle.Render("Goal:"),
			contentStyle.Render(goal), // Uses potentially updated goal
			labelStyle.Render("Return Format:"),
			contentStyle.Render(returnFormat), // Uses potentially updated returnFormat
			labelStyle.Render("Warnings:"),
			contentStyle.Render(warnings),
			labelStyle.Render("Context Dump:"),
			contentStyle.Render(contextDump),
		)

		fmt.Println(
			lipgloss.NewStyle().
				Render(sb.String()),
		)
	}
}
