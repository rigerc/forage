package ui

import (
	"fmt"
	"os"
	"strconv"

	"charm.land/lipgloss/v2"
)

var (
	boldStyle   = lipgloss.NewStyle().Bold(true)
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	cyanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type Result struct {
	Name        string
	Status      string
	Commits     string
	Files       string
	SparseLabel string
}

func LogError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s: %s\n", redStyle.Render("error"), msg)
}

func LogInfo(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", cyanStyle.Render("::"), msg)
}

func LogSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s: %s\n", greenStyle.Render("ok"), msg)
}

func LogWarn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s: %s\n", yellowStyle.Render("warn"), msg)
}

func PrintSummary(results []Result) {
	fmt.Println()
	separator := boldStyle.Render("─────────────────────────────────────────────────────────")
	fmt.Println(separator)
	fmt.Println(boldStyle.Render(" Summary"))
	fmt.Println(separator)

	fmt.Printf(
		"  %-25s %-10s %-10s %-8s %-8s\n",
		"REPO",
		"STATUS",
		"COMMITS",
		"FILES",
		"SPARSE",
	)
	fmt.Println("  ─────────────────────────────────────────────────────────")

	cloned, pulled, current, failed := 0, 0, 0, 0
	totalCommits, totalFiles := 0, 0

	for _, r := range results {
		switch r.Status {
		case "cloned":
			cloned++
		case "pulled":
			pulled++
		case "current":
			current++
		case "failed":
			failed++
		}

		if c := parseCount(r.Commits); c > 0 {
			totalCommits += c
		}
		if f := parseCount(r.Files); f > 0 {
			totalFiles += f
		}

		var statusStyle lipgloss.Style
		switch r.Status {
		case "cloned":
			statusStyle = greenStyle
		case "pulled":
			statusStyle = yellowStyle
		case "current":
			statusStyle = cyanStyle
		case "failed":
			statusStyle = redStyle
		}

		fmt.Printf(
			"  %-25s %s %-10s %-8s %-8s\n",
			r.Name,
			statusStyle.Render(fmt.Sprintf("%-10s", r.Status)),
			r.Commits,
			r.Files,
			r.SparseLabel,
		)
	}

	fmt.Println("  ─────────────────────────────────────────────────────────")
	totalStatus := fmt.Sprintf("%dc/%dp/%du/%df", cloned, pulled, current, failed)
	fmt.Printf(
		"  %s %-10s %-10s %-8s\n",
		boldStyle.Render(fmt.Sprintf("%-25s", "TOTAL")),
		totalStatus,
		strconv.Itoa(totalCommits),
		strconv.Itoa(totalFiles),
	)
	fmt.Println(separator)
}

func parseCount(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
