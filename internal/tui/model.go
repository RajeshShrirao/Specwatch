package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
)

type ActivityItem struct {
	File      string
	Status    string
	Timestamp time.Time
}

type InternalViolation struct {
	analyzer.Violation
	Timestamp time.Time
}

type Model struct {
	Activity   []ActivityItem
	Violations []InternalViolation
	TotalFiles int
	ErrorCount int
	WarnCount  int
	Latency    string
	ShowDetail bool
	Cursor     int
	Width      int
	Height     int
	Analyzing  bool

	AnimationPhase string
	AnimationFrame int
	Quitting       bool
}

type NewViolationMsg struct {
	File       string
	Violations []analyzer.Violation
	Duration   time.Duration
}

type tickMsg time.Time

// Animation phase constants
const (
	StatusDrift    = "drift"
	StatusResolved = "resolved"
	StatusClean    = "clean"

	PhaseBoot    = "boot"
	PhaseReveal  = "reveal"
	PhaseLoading = "loading"
	PhaseReady   = "ready"

	PhaseExitHold    = "exit_hold"
	PhaseExitSummary = "exit_summary"
	PhaseExitWipe    = "exit_wipe"
)

var heroLines = []string{
	"  _________                   __      __          __        __    ",
	" /   _____/_____   ____  ____/  \\    /  \\_____ _/  |_  ____|  |__ ",
	" \\_____  \\\\____ \\_/ __ \\/ ___\\   \\/\\/   /\\__  \\\\   __\\/ ___/  |  \\",
	" /        \\  |_> >  ___/ /__/ \\        /  / __ \\|  | / /_/  >   Y  \\",
	"/_______  /   __/ \\___  >___  > \\__/\\  /  (____  /__| \\___  /|___|  /",
	"        \\/|__|        \\/    \\/       \\/        \\/    /_____/      \\/ ",
}

var loadingSteps = []string{
	"Booting analyzer",
	"Loading spec parser",
	"Attaching file watcher",
	"Priming live dashboard",
}

func InitialModel() Model {
	return Model{
		Activity:       []ActivityItem{},
		Violations:     []InternalViolation{},
		AnimationPhase: PhaseBoot,
		AnimationFrame: 0,
	}
}

// handleStartupAnimation manages multi-phase startup animation
func (m Model) handleStartupAnimation() (Model, tea.Cmd) {
	m.AnimationFrame++

	switch m.AnimationPhase {
	case PhaseBoot:
		if m.AnimationFrame > 10 {
			m.AnimationPhase = PhaseReveal
			m.AnimationFrame = 0
		}
		return m, tea.Tick(70*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	case PhaseReveal:
		if m.AnimationFrame > len(heroLines[0])+8 {
			m.AnimationPhase = PhaseLoading
			m.AnimationFrame = 0
		}
		return m, tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	case PhaseLoading:
		if m.AnimationFrame > 15 {
			m.AnimationPhase = PhaseReady
			m.AnimationFrame = 0
		}
		return m, tea.Tick(75*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	case PhaseReady:
		if m.Width > 0 {
			return m, nil
		}
		return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	default:
		return m, nil
	}
}

// handleQuittingAnimation manages multi-phase shutdown animation
func (m Model) handleQuittingAnimation() (Model, tea.Cmd) {
	m.AnimationFrame++

	switch m.AnimationPhase {
	case PhaseExitHold:
		if m.AnimationFrame > 5 {
			m.AnimationPhase = PhaseExitSummary
			m.AnimationFrame = 0
		}
		return m, tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	case PhaseExitSummary:
		if m.AnimationFrame > 8 {
			m.AnimationPhase = PhaseExitWipe
			m.AnimationFrame = 0
		}
		return m, tea.Tick(85*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	case PhaseExitWipe:
		if m.AnimationFrame > 10 {
			return m, tea.Quit
		}
		return m, tea.Tick(55*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })

	default:
		return m, tea.Quit
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.Quitting {
			return m.handleQuittingAnimation()
		}
		return m.handleStartupAnimation()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		monitorWidth = m.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			m.AnimationPhase = PhaseExitHold
			m.AnimationFrame = 0
			return m, tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Violations)-1 {
				m.Cursor++
			}
		case "c":
			m.Violations = []InternalViolation{}
			m.ErrorCount = 0
			m.WarnCount = 0
			m.Cursor = 0
		case "enter":
			if len(m.Violations) > 0 {
				m.ShowDetail = !m.ShowDetail
			}
		}

	case NewViolationMsg:
		now := time.Now()
		hadViolations := false
		for _, v := range m.Violations {
			if v.File == msg.File {
				hadViolations = true
				break
			}
		}

		status := StatusClean
		if len(msg.Violations) > 0 {
			status = StatusDrift
		} else if hadViolations {
			status = StatusResolved
		}

		m.Activity = append([]ActivityItem{{
			File:      msg.File,
			Status:    status,
			Timestamp: now,
		}}, m.Activity...)
		if len(m.Activity) > 30 {
			m.Activity = m.Activity[:30]
		}

		var updatedViolations []InternalViolation
		for _, v := range m.Violations {
			if v.File != msg.File {
				updatedViolations = append(updatedViolations, v)
			}
		}
		for _, v := range msg.Violations {
			updatedViolations = append(updatedViolations, InternalViolation{
				Violation: v,
				Timestamp: now,
			})
		}

		sort.Slice(updatedViolations, func(i, j int) bool {
			if updatedViolations[i].Severity != updatedViolations[j].Severity {
				return updatedViolations[i].Severity == spec.SeverityError
			}
			return updatedViolations[i].Timestamp.After(updatedViolations[j].Timestamp)
		})
		m.Violations = updatedViolations

		m.ErrorCount = 0
		m.WarnCount = 0
		uniqueFiles := make(map[string]bool)
		for _, v := range m.Violations {
			if v.Severity == spec.SeverityError {
				m.ErrorCount++
			} else {
				m.WarnCount++
			}
			uniqueFiles[v.File] = true
		}

		seenFiles := make(map[string]bool)
		for _, a := range m.Activity {
			seenFiles[a.File] = true
		}
		m.TotalFiles = len(seenFiles)
		m.Latency = fmt.Sprintf("%dms", msg.Duration.Milliseconds())
	}

	return m, nil
}

func (m Model) View() string {
	// Startup animation phases
	switch m.AnimationPhase {
	case PhaseBoot, PhaseReveal, PhaseLoading:
		return m.renderStartupAnimation()
	}

	// Shutdown animation phases
	if m.Quitting {
		return m.renderClosingAnimation()
	}

	if m.Width == 0 {
		return renderSplash()
	}
	return m.renderMainView()
}

func (m Model) renderStartupAnimation() string {
	switch m.AnimationPhase {
	case PhaseBoot:
		return m.renderBootSequence()
	case PhaseReveal:
		return m.renderHeroReveal()
	case PhaseLoading:
		return m.renderLaunchSequence()
	default:
		return m.renderLaunchSequence()
	}
}

func (m Model) renderBootSequence() string {
	frame := m.AnimationFrame
	width := max(46, minInt(m.Width-8, 72))
	sweep := frame % width

	var rail strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i == sweep:
			rail.WriteRune('█')
		case absInt(i-sweep) <= 2:
			rail.WriteRune('▓')
		case i%8 == frame%8:
			rail.WriteRune('▒')
		default:
			rail.WriteRune('─')
		}
	}

	statuses := []string{
		"handshake terminal renderer",
		"stitching panels",
		"calibrating watch loop",
	}
	status := statuses[frame%len(statuses)]

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		StyleBootBadge.Render(" SPECWATCH "),
		StyleHeroSubtle.Render("architectural drift monitor"),
		"",
		StyleBootRail.Render(rail.String()),
		StyleFooterHint.Render("sync "+status),
	)

	return m.renderCentered(block)
}

func (m Model) renderHeroReveal() string {
	highlight := m.AnimationFrame - 6
	lines := make([]string, 0, len(heroLines))

	for _, line := range heroLines {
		segment := []rune(line)
		var builder strings.Builder
		for idx, ch := range segment {
			if idx >= m.AnimationFrame {
				builder.WriteRune(' ')
				continue
			}
			style := StyleHero
			if absInt(idx-highlight) <= 1 {
				style = StyleHeroHighlight
			}
			builder.WriteString(style.Render(string(ch)))
		}
		lines = append(lines, builder.String())
	}

	glow := StyleHeroSubtle.Render("OpenAPI rules. Live feedback. Zero ceremony.")
	block := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, lines...),
		"",
		glow,
	)

	return m.renderCentered(block)
}

func (m Model) renderLaunchSequence() string {
	progress := minInt(m.AnimationFrame+2, 15)
	barWidth := 28
	filled := progress * barWidth / 15

	bar := StyleLoadBarEmpty.Render(strings.Repeat("░", barWidth))
	if filled > 0 {
		bar = StyleLoadBarFill.Render(strings.Repeat("█", filled)) +
			StyleLoadBarEmpty.Render(strings.Repeat("░", barWidth-filled))
	}

	percent := fmt.Sprintf("%3d%%", progress*100/15)
	stepIndex := minInt((progress-1)*len(loadingSteps)/15, len(loadingSteps)-1)
	stepRows := make([]string, 0, len(loadingSteps))
	for i, step := range loadingSteps {
		prefix := "○"
		style := StyleFooterHint
		if i < stepIndex {
			prefix = "●"
			style = StyleLoadStepDone
		} else if i == stepIndex {
			prefix = "◉"
			style = StyleLoadStepActive
		}
		stepRows = append(stepRows, style.Render(prefix+" "+step))
	}

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, renderHeroStatic()...),
		"",
		StyleLoadLabel.Render("launch sequence"),
		StyleLoadMeta.Render(bar+"  "+percent),
		"",
		lipgloss.JoinVertical(lipgloss.Left, stepRows...),
	)

	return m.renderCentered(block)
}

func (m Model) renderClosingAnimation() string {
	switch m.AnimationPhase {
	case PhaseExitHold:
		return m.renderExitHold()
	case PhaseExitSummary:
		return m.renderExitSummary()
	case PhaseExitWipe:
		return m.renderExitWipe()
	default:
		m.AnimationFrame = 0
		return m.renderExitWipe()
	}
}

func (m Model) renderExitHold() string {
	tagline := StyleShutdownLabel.Render("watch ended cleanly")
	message := StyleShutdownText.Render("Persisting last session snapshot")
	pulse := strings.Repeat("•", 3+(m.AnimationFrame%3))

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		StyleBootBadge.Render(" SPECWATCH "),
		"",
		tagline,
		message,
		StyleFooterHint.Render(pulse),
	)

	return m.renderCentered(block)
}

func (m Model) renderExitSummary() string {
	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		StyleSummaryStat.Width(16).Render("FILES\n"+StyleStatValue.Render(fmt.Sprintf("%d", m.TotalFiles))),
		StyleSummaryStat.Width(16).Render("ERRORS\n"+StyleStatError.Render(fmt.Sprintf("%d", m.ErrorCount))),
		StyleSummaryStat.Width(16).Render("WARNINGS\n"+StyleStatWarning.Render(fmt.Sprintf("%d", m.WarnCount))),
	)

	lines := renderHeroStatic()
	visible := max(1, len(lines)-m.AnimationFrame/3)
	block := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, lines[:visible]...),
		"",
		StyleShutdownLabel.Render("session summary"),
		stats,
	)

	return m.renderCentered(block)
}

func (m Model) renderExitWipe() string {
	width := max(42, minInt(m.Width-8, 74))
	cut := minInt(width, m.AnimationFrame*7)
	line := StyleShutdownLine.Render(strings.Repeat("█", max(0, width-cut)))
	if cut > 0 {
		line += StyleShutdownFade.Render(strings.Repeat("░", cut))
	}

	content := []string{
		StyleShutdownLabel.Render("terminal released"),
		StyleFooterHint.Render("until next save"),
		"",
		line,
	}

	return m.renderCentered(lipgloss.JoinVertical(lipgloss.Center, content...))
}

func renderSplash() string {
	block := lipgloss.JoinVertical(
		lipgloss.Center,
		StyleBootBadge.Render(" SPECWATCH "),
		StyleHeroSubtle.Render("initializing viewport"),
	)

	return StyleBase.Render(lipgloss.Place(80, 12, lipgloss.Center, lipgloss.Center, block))
}

func (m Model) renderMainView() string {
	if m.Width < 84 || m.Height < 18 {
		return m.renderCompactView()
	}

	headerH := 1
	footerH := 1
	detailH := 0
	if m.ShowDetail && len(m.Violations) > 0 {
		detailH = 9
	}

	mainH := m.Height - headerH - footerH - detailH - 3
	if mainH < 5 {
		mainH = 5
	}

	leftW := max(25, m.Width/4)
	rightW := 22
	centerW := m.Width - leftW - rightW - 3

	// Header
	statusDot := StyleStatusDot.Render("●")
	if m.ErrorCount > 0 {
		statusDot = StyleStatusDotError.Render("●")
	}
	headerContent := fmt.Sprintf(" %s %s %s %s",
		StyleLogo.Render("⚡ specwatch"),
		StyleFooterHint.Render("│"),
		statusDot+" "+StyleFooterHint.Render(m.statusLabel()),
		StyleFooterLatency.Render(m.Latency),
	)
	header := StyleHeader.Width(m.Width).Render(headerContent)

	// Left Panel - Activity
	leftB := new(strings.Builder)
	leftB.WriteString(StyleStatLabel.Render(" ACTIVITY ") + "\n")
	leftB.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

	maxAct := mainH - 3
	visAct := m.Activity
	if len(visAct) > maxAct {
		visAct = visAct[:maxAct]
	}

	for _, item := range visAct {
		dot := m.renderActivityStatus(item.Status)
		filename := truncate(filepath.Base(item.File), leftW-18)
		elapsed := time.Since(item.Timestamp).Round(time.Second).String()
		leftB.WriteString(fmt.Sprintf("%s %-13s %s\n", dot, filename, StyleActivityTime.Render(elapsed)))
	}
	if len(visAct) == 0 {
		leftB.WriteString(StyleEmptyMuted.Render("  Waiting for file activity\n"))
		leftB.WriteString(StyleFooterHint.Render("  Save a tracked file to start"))
	}
	leftPanel := StylePanel.Width(leftW).Height(mainH).Render(leftB.String())

	// Center Panel - Violations
	centerB := new(strings.Builder)
	centerB.WriteString(StyleStatLabel.Render(" VIOLATIONS ") + "\n")
	centerB.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

	if len(m.Violations) == 0 {
		centerB.WriteString(StyleEmptyTitle.Render("  No active drift"))
		centerB.WriteString("\n")
		if len(m.Activity) > 0 && m.Activity[0].Status == StatusResolved {
			centerB.WriteString(StyleActivityResolved.Render("  Latest issue resolved successfully"))
		} else {
			centerB.WriteString(StyleFooterHint.Render("  Save a file to analyze"))
		}
		centerB.WriteString("\n\n")
		centerB.WriteString(StyleEmptyMuted.Render("  Specwatch will replace stale findings on the next event"))
	} else {
		maxViol := mainH - 3
		start := 0
		if m.Cursor >= maxViol {
			start = m.Cursor - maxViol + 1
		}
		end := start + maxViol
		if end > len(m.Violations) {
			end = len(m.Violations)
		}

		for i := start; i < end; i++ {
			v := m.Violations[i]
			isSel := i == m.Cursor

			loc := fmt.Sprintf("%s:%d", truncate(filepath.Base(v.File), 15), v.Line)
			rule := truncate(v.Rule, centerW-22)

			if isSel {
				sevI := "✗"
				if v.Severity == spec.SeverityWarning {
					sevI = "⚠"
				}
				line := StyleViolationSelected.Render(fmt.Sprintf(" %s %s %s", sevI, loc, rule))
				centerB.WriteString(line + "\n")
			} else {
				sevI := StyleViolationError.Render("✗")
				if v.Severity == spec.SeverityWarning {
					sevI = StyleViolationWarning.Render("⚠")
				}
				line := fmt.Sprintf(" %s %s %s", sevI, loc, StyleStatLabel.Render(rule))
				centerB.WriteString(line + "\n")
			}
		}
	}
	centerPanel := StylePanel.Width(centerW).Height(mainH).Render(centerB.String())

	// Right Panel - Stats
	rightB := new(strings.Builder)
	rightB.WriteString(StyleStatLabel.Render(" STATS ") + "\n")
	rightB.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n\n")

	rightB.WriteString(StyleStatLabel.Render(" Files    "))
	rightB.WriteString(StyleStatValue.Render(fmt.Sprintf("%d", m.TotalFiles)) + "\n")

	rightB.WriteString(StyleStatLabel.Render(" Errors   "))
	if m.ErrorCount > 0 {
		rightB.WriteString(StyleStatError.Render(fmt.Sprintf("%d", m.ErrorCount)) + "\n")
	} else {
		rightB.WriteString(StyleStatSuccess.Render("0") + "\n")
	}

	rightB.WriteString(StyleStatLabel.Render(" Warnings "))
	if m.WarnCount > 0 {
		rightB.WriteString(StyleStatWarning.Render(fmt.Sprintf("%d", m.WarnCount)) + "\n")
	} else {
		rightB.WriteString(StyleStatSuccess.Render("0") + "\n")
	}

	rightB.WriteString("\n")
	rightB.WriteString(StyleStatLabel.Render(" State    "))
	rightB.WriteString(StyleHealthBadge.Render(strings.ToUpper(m.statusLabel())) + "\n")
	rightPanel := StylePanel.Width(rightW).Height(mainH).Render(rightB.String())

	// Detail Panel
	detailView := ""
	if m.ShowDetail && len(m.Violations) > 0 && m.Cursor < len(m.Violations) {
		v := m.Violations[m.Cursor]
		sevL := StyleViolationError.Render(" ERROR ")
		if v.Severity == spec.SeverityWarning {
			sevL = StyleViolationWarning.Render(" WARNING ")
		}

		detContent := fmt.Sprintf(
			"%s %s  %s\n\n%s %s\n%s %s\n%s %s\n%s\n%s %s",
			StyleDetailKey.Render("File:"),
			StyleDetailValue.Render(v.File),
			sevL,
			StyleDetailKey.Render("Line:"),
			StyleDetailValue.Render(fmt.Sprintf("%d", v.Line)),
			StyleDetailKey.Render("Rule:"),
			StyleDetailValue.Render(v.Rule),
			StyleDetailKey.Render("Severity:"),
			StyleDetailValue.Render(string(v.Severity)),
			StyleDetailCode.Render(truncate(v.Excerpt, m.Width-20)),
			StyleDetailKey.Render("Fix:"),
			StyleDetailSuggestion.Render(truncate(v.Suggestion, m.Width-20)),
		)
		detailView = StyleDetail.Width(m.Width - 2).Height(detailH - 1).Render(detContent)
	}

	// Footer
	footerC := fmt.Sprintf(
		" %s %s %s %s %s %s %s %s %s",
		StyleFooterKey.Render("↑↓"),
		StyleFooterHint.Render("Nav"),
		StyleFooterKey.Render("ENT"),
		StyleFooterHint.Render("Detail"),
		StyleFooterKey.Render("C"),
		StyleFooterHint.Render("Clear"),
		StyleFooterKey.Render("Q"),
		StyleFooterHint.Render("Quit"),
		StyleFooterLatency.Render(" ⚡ "+m.Latency+" "),
	)
	footer := StyleFooter.Width(m.Width).Render(footerC)

	// Combine
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)

	return StyleBase.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			panels,
			detailView,
			footer,
		),
	)
}

func (m Model) renderCompactView() string {
	header := StyleHeader.Width(max(32, m.Width)).Render(fmt.Sprintf(
		" %s %s %s",
		StyleLogo.Render("⚡ specwatch"),
		StyleFooterHint.Render("│"),
		StyleFooterHint.Render(m.statusLabel()),
	))

	latest := StyleEmptyMuted.Render("No file activity yet")
	if len(m.Activity) > 0 {
		item := m.Activity[0]
		latest = fmt.Sprintf(
			"%s %s %s",
			m.renderActivityStatus(item.Status),
			truncate(filepath.Base(item.File), max(12, m.Width-24)),
			StyleActivityTime.Render(time.Since(item.Timestamp).Round(time.Second).String()),
		)
	}

	summary := lipgloss.JoinVertical(
		lipgloss.Left,
		StyleCompactCard.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				StyleEmptyTitle.Render("Live status"),
				StyleHealthBadge.Render(strings.ToUpper(m.statusLabel())),
				"",
				StyleStatLabel.Render("Latest activity"),
				latest,
			),
		),
		"",
		StyleCompactCard.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				StyleStatLabel.Render("Files"),
				StyleStatValue.Render(fmt.Sprintf("%d", m.TotalFiles)),
				StyleStatLabel.Render("Errors"),
				StyleStatError.Render(fmt.Sprintf("%d", m.ErrorCount)),
				StyleStatLabel.Render("Warnings"),
				StyleStatWarning.Render(fmt.Sprintf("%d", m.WarnCount)),
			),
		),
	)

	if len(m.Violations) > 0 {
		v := m.Violations[m.Cursor]
		summary = lipgloss.JoinVertical(
			lipgloss.Left,
			summary,
			"",
			StyleCompactCard.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					StyleStatLabel.Render("Current drift"),
					StyleDetailValue.Render(filepath.Base(v.File)),
					StyleFooterHint.Render(fmt.Sprintf("line %d • %s", v.Line, v.Rule)),
					StyleEmptyMuted.Render(truncate(v.Suggestion, max(24, m.Width-12))),
				),
			),
		)
	}

	footer := StyleFooter.Width(max(32, m.Width)).Render(
		fmt.Sprintf(" %s %s %s %s", StyleFooterKey.Render("Q"), StyleFooterHint.Render("Quit"), StyleFooterKey.Render("ENT"), StyleFooterHint.Render("Detail")),
	)

	return StyleBase.Render(lipgloss.JoinVertical(lipgloss.Left, header, summary, footer))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func renderHeroStatic() []string {
	lines := make([]string, 0, len(heroLines))
	for _, line := range heroLines {
		lines = append(lines, StyleHero.Render(line))
	}
	return lines
}

func (m Model) renderCentered(content string) string {
	width := m.Width
	height := m.Height
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 28
	}
	return StyleBase.Render(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content))
}

func (m Model) renderActivityStatus(status string) string {
	switch status {
	case StatusDrift:
		return StyleActivityError.Render("● drift")
	case StatusResolved:
		return StyleActivityResolved.Render("● fixed")
	default:
		return StyleActivityClean.Render("● clean")
	}
}

func (m Model) statusLabel() string {
	switch {
	case m.ErrorCount > 0:
		return "drift detected"
	case m.WarnCount > 0:
		return "watching warnings"
	case len(m.Activity) > 0 && m.Activity[0].Status == StatusResolved:
		return "drift resolved"
	default:
		return "watching stable"
	}
}
