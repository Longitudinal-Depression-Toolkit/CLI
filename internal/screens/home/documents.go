package home

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"

	"ldt-toolkit-cli/internal/shared/components"
)

func IsGoNativeAction(path []string) bool {
	if len(path) != 1 {
		return false
	}
	switch path[0] {
	case "list_authors", "list_inspiration_from", "toolkit_tutorial":
		return true
	default:
		return false
	}
}

func RunConfiguredAuthors() error {
	markdown := markdownForLabelledEntries(currentHomeConfig.Authors)
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		components.PrintBlankLine()
		components.PrintLine(markdown)
		return nil
	}
	return components.ShowMarkdownDocument("Authors", markdown)
}

var urlPattern = regexp.MustCompile(`https?://[^\s)]+`)

func markdownForLabelledEntries(entries []LabelledEntry) string {
	lines := make([]string, 0, len(entries)+2)
	if len(entries) == 0 {
		lines = append(lines, "No entries configured.")
		return strings.Join(lines, "\n")
	}
	for _, entry := range entries {
		lines = append(
			lines,
			fmt.Sprintf(
				"- **%s)** %s",
				strings.TrimSpace(entry.Label),
				linkifyURLs(strings.TrimSpace(entry.Text)),
			),
		)
	}
	lines = append(lines, "", "_Scroll with mouse wheel or arrow keys. Press `q` to close._")
	return strings.Join(lines, "\n")
}

func RunConfiguredInspirations() error {
	markdown := markdownForLabelledEntries(currentHomeConfig.Inspirations)
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		components.PrintBlankLine()
		components.PrintLine(markdown)
		return nil
	}
	return components.ShowMarkdownDocument("Inspiration From", markdown)
}

func RunConfiguredTutorial() error {
	markdown := tutorialMarkdownFromConfig(currentHomeConfig.Tutorial)
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		components.PrintBlankLine()
		components.PrintLine(markdown)
		return nil
	}
	return components.ShowMarkdownDocument("LDT Toolkit Tutorial", markdown)
}

func tutorialMarkdownFromConfig(steps []TutorialStep) string {
	lines := []string{
		"Stage-by-stage guide for running Longitudinal Depression Toolkit workflows.",
		"",
	}

	if len(steps) == 0 {
		lines = append(lines, "No tutorial steps configured yet.")
		return strings.Join(lines, "\n")
	}

	for index, step := range steps {
		lines = append(
			lines,
			fmt.Sprintf("## %d. %s", index+1, step.Title),
			"",
			"### Goal",
			step.Goal,
			"",
			"### Outcome",
			step.Outcome,
			"",
		)
		if step.Recommendation != "" {
			lines = append(lines, "### Recommendation", step.Recommendation, "")
		}
		lines = append(lines, "---", "")
	}

	lines = append(lines, "_Scroll with mouse wheel or arrow keys. Press `q` to close._")
	return strings.Join(lines, "\n")
}

func linkifyURLs(value string) string {
	return urlPattern.ReplaceAllStringFunc(value, func(url string) string {
		return fmt.Sprintf("[%s](%s)", url, url)
	})
}
