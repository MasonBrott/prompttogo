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

var (
	goal         string
	returnFormat string
	warnings     string
	contextDump  string
	confirm      bool
)

func main() {

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Goal").
				CharLimit(500).
				Placeholder("e.g., Write a function to reverse a string").
				Value(&goal),
			huh.NewText().
				Title("Return Format").
				CharLimit(500).
				Placeholder("e.g., Just the code, no explanation").
				Value(&returnFormat),
			huh.NewText().
				Title("Warnings").
				CharLimit(500).
				Placeholder("e.g., Do not use the built-in reverse function").
				Value(&warnings),
			huh.NewText().
				Title("Context Dump").
				CharLimit(500).
				Placeholder("e.g., The function signature is func reverse(s string) string").
				Value(&contextDump),
			huh.NewConfirm().
				Title("Are you ready to generate your prompt?").
				Affirmative("Yes!").
				Negative("No!").
				Value(&confirm),
		),
	)

	err := form.Run()
	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	preparePrompt := func() {
		time.Sleep(2 * time.Second)
	}

	_ = spinner.New().Title("Preparing your prompt...").Action(preparePrompt).Run()

	{
		var sb strings.Builder
		keyword := func(s string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
		}

		fmt.Fprintf(&sb,
			"Goal:\n%s\n\nReturn Format:\n%s\n\nWarnings:\n%s\n\nContext Dump:\n%s\n\n",
			keyword(goal),
			keyword(returnFormat),
			keyword(warnings),
			keyword(contextDump),
		)

		fmt.Println(
			lipgloss.NewStyle().
				Render(sb.String()),
		)
	}
}
