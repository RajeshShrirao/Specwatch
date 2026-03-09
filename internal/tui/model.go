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
	Clean     bool
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

func InitialModel() Model {
	return Model{
		Activity:       []ActivityItem{},
		Violations:     []InternalViolation{},
		AnimationPhase: "start",
		AnimationFrame: 0,
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
			m.AnimationFrame++
			if m.AnimationFrame > 15 {
				return m, tea.Quit
			}
			return m, tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
		}

		if m.AnimationPhase == "start" {
			m.AnimationFrame++
			if m.AnimationFrame > 30 {
				m.AnimationPhase = "ready"
				m.AnimationFrame = 0
			}
			return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
		}

		if m.AnimationPhase == "ready" && m.Width > 0 {
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		monitorWidth = m.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			m.AnimationPhase = "quit"
			m.AnimationFrame = 0
			return m, tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
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
		m.Activity = append([]ActivityItem{{
			File:      msg.File,
			Clean:     len(msg.Violations) == 0,
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
	if m.AnimationPhase == "start" {
		return m.renderStartupAnimation()
	}
	if m.AnimationPhase == "quit" {
		return m.renderClosingAnimation()
	}
	if m.Width == 0 {
		return renderSplash()
	}
	return m.renderMainView()
}

func (m Model) renderStartupAnimation() string {
	frames := []string{"⚡", "⚡ s", "⚡ sp", "⚡ spe", "⚡ spec", "⚡ specw", "⚡ specwa", "⚡ specwat", "⚡ specwatc", "⚡ specwatch"}

	frame := m.AnimationFrame
	if frame >= len(frames) {
		frame = len(frames) - 1
	}
	if frame < 0 {
		frame = 0
	}

	text := frames[frame]
	cursor := "▌"
	if frame%2 == 0 {
		cursor = "█"
	}

	centerX := 35
	padding := strings.Repeat(" ", centerX-len(text)/2)

	splash := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#58a6ff")).
		Bold(true).
		Render(padding + text + cursor)

	barW := 25
	prog := float64(frame) / float64(30)
	filled := int(float64(barW) * prog)
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", barW-filled)

	loading := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3fb950")).
		Render(bar)

	return StyleBase.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			"\n\n\n\n\n", splash, "\n", loading,
			"\n\n", StyleFooterHint.Render("Initializing..."),
		),
	)
}

func (m Model) renderClosingAnimation() string {
	frames := []string{
		"⚡ specwatch", "⚡ specwatc", "⚡ specwat", "⚡ specwa",
		"⚡ specw", "⚡ spec", "⚡ spe", "⚡ sp", "⚡ s", "⚡", "",
	}

	frame := m.AnimationFrame
	if frame >= len(frames) {
		frame = len(frames) - 1
	}

	text := frames[frame]
	centerX := 35
	padding := strings.Repeat(" ", centerX-len(text)/2)

	splash := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f85149")).
		Bold(true).
		Render(padding + text)

	spinner := []string{"○", "◔", "◑", "◕", "◉", "◕", "◑", "◔"}
	sp := spinner[frame%len(spinner)]

	return StyleBase.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			"\n\n\n\n\n", splash, "\n", sp,
			"\n\n", StyleFooterHint.Render("Shutting down..."),
		),
	)
}

func renderSplash() string {
	padding := strings.Repeat(" ", 27)
	splash := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#58a6ff")).
		Bold(true).
		Render(padding + "⚡ specwatch")

	return StyleBase.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			"\n\n\n\n\n", splash, "\n\n\n\n",
		),
	)
}

func (m Model) renderMainView() string {
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
		statusDot+" "+StyleFooterHint.Render("watching"),
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
		dot := StyleActivityClean.Render("●")
		if !item.Clean {
			dot = StyleActivityError.Render("●")
		}
		filename := truncate(filepath.Base(item.File), leftW-15)
		elapsed := time.Since(item.Timestamp).Round(time.Second).String()
		leftB.WriteString(fmt.Sprintf("%s %-12s %s\n", dot, filename, StyleActivityTime.Render(elapsed)))
	}
	leftPanel := StylePanel.Width(leftW).Height(mainH).Render(leftB.String())

	// Center Panel - Violations
	centerB := new(strings.Builder)
	centerB.WriteString(StyleStatLabel.Render(" VIOLATIONS ") + "\n")
	centerB.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

	if len(m.Violations) == 0 {
		centerB.WriteString(StyleFooterHint.Render("  No violations detected"))
		centerB.WriteString("\n\n")
		centerB.WriteString(StyleFooterHint.Render("  Save a file to analyze"))
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
			"%s %s  %s\n\n%s %s\n%s %d\n%s %s\n%s\n%s %s",
			StyleDetailKey.Render("File:"),
			StyleDetailValue.Render(v.File),
			sevL,
			StyleDetailKey.Render("Line:"),
			StyleDetailValue.Render(fmt.Sprintf("%d", v.Line)),
			StyleDetailKey.Render("Rule:"),
			v.Line,
			StyleDetailKey.Render("Category:"),
			StyleDetailValue.Render(v.Rule),
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
