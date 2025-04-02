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
	goal             string
	returnFormat     string
	warnings         string // User's original warnings
	contextDump      string
	selectedWarnings []string // Warnings selected from suggestions
	confirm          bool
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

// getEnrichmentSuggestions provides refined suggestions for goal, format, and warnings.
func getEnrichmentSuggestions(archetype string, originalGoal string, originalFormat string) (suggestedGoal string, suggestedFormats []huh.Option[string], suggestedWarnings []huh.Option[string]) {
	switch archetype {
	case ArchetypeSummarization:
		suggestedGoal = "Summarize the key requirements and obligations mentioned in the provided documents."
		suggestedFormats = []huh.Option[string]{
			huh.NewOption("Bulleted list of key points", "Bulleted list of key points"),
			huh.NewOption("Concise paragraph overview", "Concise paragraph overview"),
		}
		suggestedWarnings = []huh.Option[string]{
			huh.NewOption("Focus only on actionable requirements", "Focus on requirements"),
			huh.NewOption("Avoid technical jargon where possible", "Avoid jargon"),
		}
		// Add original format as an option if it's different and not empty
		isOriginalAlreadySuggested := false
		for _, opt := range suggestedFormats {
			if opt.Value == originalFormat {
				isOriginalAlreadySuggested = true
				break
			}
		}
		if !isOriginalAlreadySuggested {
			if originalFormat != "" {
				suggestedFormats = append(suggestedFormats, huh.NewOption("Keep: "+originalFormat, originalFormat))
			} else {
				suggestedFormats = append(suggestedFormats, huh.NewOption("Keep original (empty)", ""))
			}
		}

	case ArchetypeQA:
		// Prepend standard framing to the user's original question/goal
		suggestedGoal = fmt.Sprintf("Based *only* on the provided documents, answer the question: %s", originalGoal)
		suggestedFormats = []huh.Option[string]{
			huh.NewOption("Direct answer", "Direct answer"),
			huh.NewOption("Answer with citations to relevant sections", "Answer with citations"),
			huh.NewOption("Extract relevant quotes supporting the answer", "Extract relevant quotes"),
		}
		suggestedWarnings = []huh.Option[string]{
			huh.NewOption("Do not infer information not explicitly present", "Do not infer"),
			huh.NewOption("Cite the source section(s) for the answer", "Cite sources"),
			huh.NewOption("If the answer is not found, state that clearly", "State if not found"),
		}
		// Add original format as an option if it's different and not empty
		isOriginalAlreadySuggested := false
		for _, opt := range suggestedFormats {
			if opt.Value == originalFormat {
				isOriginalAlreadySuggested = true
				break
			}
		}
		if !isOriginalAlreadySuggested {
			if originalFormat != "" {
				suggestedFormats = append(suggestedFormats, huh.NewOption("Keep: "+originalFormat, originalFormat))
			} else {
				suggestedFormats = append(suggestedFormats, huh.NewOption("Keep original (empty)", ""))
			}
		}

	default:
		// No suggestions for unknown archetype
		return originalGoal, nil, nil
	}

	// Ensure the original format is the default selection if it exists in the options
	// If not, the first suggestion becomes the default. Huh handles default selection.
	originalIndex := -1
	for i, opt := range suggestedFormats {
		if opt.Value == originalFormat {
			originalIndex = i
			break
		}
	}
	if originalIndex > 0 {
		originalOpt := suggestedFormats[originalIndex]
		suggestedFormats = append(suggestedFormats[:originalIndex], suggestedFormats[originalIndex+1:]...)
		suggestedFormats = append([]huh.Option[string]{originalOpt}, suggestedFormats...)
	}

	return suggestedGoal, suggestedFormats, suggestedWarnings
}

func main() {

	// Outer loop to allow restarting the process
	for {
		// Reset variables at the start of each loop iteration
		// (goal, returnFormat, warnings are bound to form fields, will be overwritten)
		selectedWarnings = []string{} // Reset selected warnings
		confirm = false               // Reset confirmation

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
					Value(&warnings), // Captures user's manual warning input
				huh.NewText().
					Title("Context Dump").
					Placeholder("e.g., Paste relevant sections of compliance docs here").
					CharLimit(500).
					Value(&contextDump),
			),
		)

		err := initialForm.Run()
		if err != nil {
			if err == huh.ErrUserAborted {
				fmt.Println("\nOperation cancelled by user.")
				os.Exit(0)
			}
			fmt.Println("Uh oh:", err)
			os.Exit(1)
		}
		originalGoal := goal
		originalFormat := returnFormat

		// --- Archetype Detection and Guidance ---
		detectedArchetype := detectArchetype(goal)
		if detectedArchetype != ArchetypeUnknown {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("\nPrompt Guidance:"))
			guidance := getGuidance(detectedArchetype)
			for _, tip := range guidance {
				fmt.Println(lipgloss.NewStyle().Faint(true).Render("- " + tip))
			}
			fmt.Println()

			// --- Enrichment Step ---
			fmt.Println(lipgloss.NewStyle().Bold(true).Render("Step 2: Refine Prompt (Optional)"))
			suggestedGoal, suggestedFormats, suggestedWarningsOptions := getEnrichmentSuggestions(detectedArchetype, originalGoal, originalFormat)

			goal = suggestedGoal // Pre-fill refined goal

			enrichmentForm := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().Title("Detected Intent: "+detectedArchetype).Description("We detected a potential intent. You can refine the Goal, Return Format, and add common Warnings below."),
					huh.NewText().
						Title("Refined Goal").
						Description("Suggested goal based on detection. Edit as needed.").
						Value(&goal),
					huh.NewSelect[string]().
						Title("Suggested Return Format").
						Options(suggestedFormats...). // Default set by getEnrichmentSuggestions
						Value(&returnFormat),
					huh.NewMultiSelect[string]().
						Title("Add Common Warnings (Optional)").
						Options(suggestedWarningsOptions...).
						Value(&selectedWarnings),
				),
			)

			err = enrichmentForm.Run()
			if err != nil {
				if err == huh.ErrUserAborted {
					fmt.Println("\nOperation cancelled by user.")
					os.Exit(0)
				}
				fmt.Println("Error during enrichment:", err)
				os.Exit(1)
			}
		} else {
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
		if err != nil {
			if err == huh.ErrUserAborted {
				fmt.Println("\nOperation cancelled by user.")
				os.Exit(0)
			}
			fmt.Println("Confirmation error:", err)
			os.Exit(1) // Exit on other confirmation errors
		}

		if confirm {
			break // Exit the loop and proceed to generation
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("\nRestarting prompt generation..."))
			time.Sleep(1 * time.Second) // Brief pause before restarting
			// Loop will continue automatically
		}
	} // End of the main loop

	// --- Spinner and Final Output ---
	preparePrompt := func() {
		time.Sleep(1 * time.Second)
	}

	_ = spinner.New().Title("Preparing your prompt...").Action(preparePrompt).Run()

	{
		var sb strings.Builder
		labelStyle := lipgloss.NewStyle().Bold(true)
		contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

		// Combine original warnings and selected suggested warnings
		finalWarnings := warnings // Start with user's original input
		if len(selectedWarnings) > 0 {
			if finalWarnings != "" {
				finalWarnings += "\n" // Add newline if original warnings exist
			}
			// Append selected warnings, prefixing each with a bullet or similar
			for _, sw := range selectedWarnings {
				finalWarnings += "- " + sw + "\n"
			}
			finalWarnings = strings.TrimSpace(finalWarnings) // Clean up trailing newline
		}

		fmt.Fprintf(&sb,
			"%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n",
			labelStyle.Render("Goal:"),
			contentStyle.Render(goal),
			labelStyle.Render("Return Format:"),
			contentStyle.Render(returnFormat),
			labelStyle.Render("Warnings:"),
			contentStyle.Render(finalWarnings), // Use the combined warnings
			labelStyle.Render("Context Dump:"),
			contentStyle.Render(contextDump),
		)

		fmt.Println(
			lipgloss.NewStyle().
				Render(sb.String()),
		)
	}
}
