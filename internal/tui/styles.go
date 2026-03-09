package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#25a065")
	ColorError     = lipgloss.Color("#E84855")
	ColorWarning   = lipgloss.Color("#FF9B54")
	ColorGrey      = lipgloss.Color("#5A5A5A")
	ColorBg        = lipgloss.Color("#1A1A1A")
	ColorSubtle    = lipgloss.Color("#333333")

	// Styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(ColorPrimary).
			Padding(0, 1)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StyleFileClean = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleFileViolation = lipgloss.NewStyle().
				Foreground(ColorError)

	StyleViolationItem = lipgloss.NewStyle().
				PaddingLeft(2)

	StyleStats = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorSubtle).
			PaddingLeft(2)

	StyleDetail = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorSubtle).
			Padding(1, 0).
			MarginTop(1)

	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2)

	StyleStatusWatching = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	StyleTime = lipgloss.NewStyle().
			Foreground(ColorGrey)

	StyleLineNum = lipgloss.NewStyle().
			Foreground(ColorGrey)

	StyleRule = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Italic(true)
)
