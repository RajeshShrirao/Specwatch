package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorBg      = lipgloss.Color("#0a0a0a")
	ColorBorder  = lipgloss.Color("#222222")
	ColorError   = lipgloss.Color("#F75F4F")
	ColorWarning = lipgloss.Color("#F7A94F")
	ColorSuccess = lipgloss.Color("#2DCC8F")
	ColorMuted   = lipgloss.Color("#555555")
	ColorAccent  = lipgloss.Color("#4F8EF7")

	// Base Styles
	StyleMain = lipgloss.NewStyle().
			Background(ColorBg).
			Foreground(lipgloss.Color("#FAFAFA"))

	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	StylePanelActive = StylePanel.Copy().
				BorderForeground(ColorAccent)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			MarginBottom(1)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorError)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleAccent = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// Panel Specifics
	StyleActivityItem = lipgloss.NewStyle().
				Height(1)

	StyleViolationItem = lipgloss.NewStyle().
				PaddingLeft(1)

	StyleViolationSelected = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1a1a")).
				Foreground(lipgloss.Color("#ffffff"))

	StyleStatValue = lipgloss.NewStyle().
			Bold(true)

	StyleDetailKey = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(15)

	StyleDetailValue = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	StyleFooterHint = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleFooterKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	StyleLatency = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)
