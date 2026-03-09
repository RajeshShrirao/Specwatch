package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - Modern dark theme
	ColorBg          = lipgloss.Color("#0d1117")
	ColorBgAlt       = lipgloss.Color("#161b22")
	ColorBorder      = lipgloss.Color("#30363d")
	ColorBorderLight = lipgloss.Color("#484f58")
	ColorError       = lipgloss.Color("#f85149")
	ColorErrorBg     = lipgloss.Color("#3d1d20")
	ColorWarning     = lipgloss.Color("#d29922")
	ColorWarningBg   = lipgloss.Color("#3d2e00")
	ColorSuccess     = lipgloss.Color("#3fb950")
	ColorSuccessBg   = lipgloss.Color("#1d3a2a")
	ColorMuted       = lipgloss.Color("#8b949e")
	ColorAccent      = lipgloss.Color("#58a6ff")
	ColorAccentBg    = lipgloss.Color("#1f2a37")
	ColorText        = lipgloss.Color("#e6edf3")
	ColorTextDim     = lipgloss.Color("#8b949e")

	// Base Styles
	StyleBase = lipgloss.NewStyle().
			Background(ColorBg).
			Foreground(ColorText)

	// Header
	StyleHeader = lipgloss.NewStyle().
			Background(ColorAccentBg).
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1).
			Width(monitorWidth)

	StyleLogo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a371f7")).
			Bold(true)

	StyleStatusDot = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleStatusDotError = lipgloss.NewStyle().
				Foreground(ColorError)

	// Panels
	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBgAlt)

	StylePanelHeader = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true).
				Padding(0, 1).
				Background(ColorBorder).
				Width(monitorWidth - 2)

	// Activity Items
	StyleActivityClean = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	StyleActivityError = lipgloss.NewStyle().
				Foreground(ColorError)

	StyleActivityTime = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)

	// Violations
	StyleViolationError = lipgloss.NewStyle().
				Foreground(ColorError).
				Background(ColorErrorBg).
				Padding(0, 1)

	StyleViolationWarning = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Background(ColorWarningBg).
				Padding(0, 1)

	StyleViolationSelected = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorAccentBg).
				Padding(0, 1).
				Bold(true)

	StyleViolationPrefix = lipgloss.NewStyle().
				Bold(true)

	// Stats
	StyleStatLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true)

	StyleStatValue = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true).
			PaddingLeft(1)

	StyleStatError = StyleStatValue.Copy().
			Foreground(ColorError)

	StyleStatWarning = StyleStatValue.Copy().
				Foreground(ColorWarning)

	StyleStatSuccess = StyleStatValue.Copy().
				Foreground(ColorSuccess)

	// Detail Panel
	StyleDetail = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorderLight).
			Background(ColorBgAlt).
			Padding(1, 2)

	StyleDetailKey = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true).
			Width(12)

	StyleDetailValue = lipgloss.NewStyle().
				Foreground(ColorText)

	StyleDetailCode = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Background(ColorAccentBg).
			Padding(0, 1)

	StyleDetailSuggestion = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Background(ColorSuccessBg).
				Padding(0, 1)

	// Footer
	StyleFooter = lipgloss.NewStyle().
			Background(ColorBgAlt).
			Foreground(ColorMuted).
			Padding(0, 1)

	StyleFooterKey = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleFooterHint = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleFooterLatency = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Italic(true).
				Bold(true)
)

var monitorWidth int
