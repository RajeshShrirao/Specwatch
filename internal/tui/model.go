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
}

type NewViolationMsg struct {
	File       string
	Violations []analyzer.Violation
	Duration   time.Duration
}

func InitialModel() Model {
	return Model{
		Activity:   []ActivityItem{},
		Violations: []InternalViolation{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		monitorWidth = m.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
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

		// Update Violations
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

		// Sort: Severity (Error > Warning), then Recency
		sort.Slice(updatedViolations, func(i, j int) bool {
			if updatedViolations[i].Severity != updatedViolations[j].Severity {
				return updatedViolations[i].Severity == spec.SeverityError
			}
			return updatedViolations[i].Timestamp.After(updatedViolations[j].Timestamp)
		})
		m.Violations = updatedViolations

		// Stats
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
	if m.Width == 0 || m.Height == 0 {
		return "Initializing specwatch..."
	}

	// Layout dimensions
	headerHeight := 1
	footerHeight := 1
	detailHeight := 0
	if m.ShowDetail && len(m.Violations) > 0 {
		detailHeight = 9
	}

	mainHeight := m.Height - headerHeight - footerHeight - detailHeight - 3
	if mainHeight < 5 {
		mainHeight = 5
	}

	// Panel widths
	leftW := max(25, m.Width/4)
	rightW := 22
	centerW := m.Width - leftW - rightW - 3

	// ========== HEADER ==========
	headerStatus := StyleStatusDot.Render("● watching")
	if m.ErrorCount > 0 {
		headerStatus = StyleStatusDotError.Render("● errors detected")
	}

	headerContent := fmt.Sprintf(" %s %s %s %s",
		StyleLogo.Render("⚡ specwatch"),
		StyleFooterHint.Render("│"),
		headerStatus,
		StyleFooterLatency.Render(m.Latency),
	)
	header := StyleHeader.Width(m.Width).Render(headerContent)

	// ========== LEFT PANEL: Activity ==========
	leftContent := new(strings.Builder)
	leftContent.WriteString(StyleStatLabel.Render(" ACTIVITY ") + "\n")
	leftContent.WriteString(strings.Repeat("─", leftW-2) + "\n")

	maxActivity := mainHeight - 3
	visibleActivity := m.Activity
	if len(visibleActivity) > maxActivity {
		visibleActivity = visibleActivity[:maxActivity]
	}

	for _, item := range visibleActivity {
		dot := StyleActivityClean.Render("●")
		if !item.Clean {
			dot = StyleActivityError.Render("●")
		}
		filename := truncate(filepath.Base(item.File), leftW-15)
		elapsed := time.Since(item.Timestamp).Round(time.Second).String()
		leftContent.WriteString(fmt.Sprintf("%s %-12s %s\n", dot, filename, StyleActivityTime.Render(elapsed)))
	}
	leftPanel := StylePanel.Width(leftW).Height(mainHeight).Render(leftContent.String())

	// ========== CENTER PANEL: Violations ==========
	centerContent := new(strings.Builder)
	centerContent.WriteString(StyleStatLabel.Render(" VIOLATIONS ") + "\n")
	centerContent.WriteString(strings.Repeat("─", centerW-2) + "\n")

	if len(m.Violations) == 0 {
		centerContent.WriteString(StyleStatLabel.Render("  No violations detected"))
		centerContent.WriteString("\n")
		centerContent.WriteString(StyleFooterHint.Render("  Save a file to analyze"))
	} else {
		maxViolations := mainHeight - 3
		start := 0
		if m.Cursor >= maxViolations {
			start = m.Cursor - maxViolations + 1
		}
		end := start + maxViolations
		if end > len(m.Violations) {
			end = len(m.Violations)
		}

		for i := start; i < end; i++ {
			v := m.Violations[i]
			isSelected := i == m.Cursor

			loc := fmt.Sprintf("%s:%d", truncate(filepath.Base(v.File), 15), v.Line)
			rule := truncate(v.Rule, centerW-22)

			if isSelected {
				sevIcon := "✗"
				if v.Severity == spec.SeverityWarning {
					sevIcon = "⚠"
				}
				line := StyleViolationSelected.Render(fmt.Sprintf(" %s %s %s", sevIcon, loc, rule))
				centerContent.WriteString(line + "\n")
			} else {
				sevIcon := StyleViolationError.Render("✗")
				if v.Severity == spec.SeverityWarning {
					sevIcon = StyleViolationWarning.Render("⚠")
				}
				line := fmt.Sprintf(" %s %s %s", sevIcon, loc, StyleStatLabel.Render(rule))
				centerContent.WriteString(line + "\n")
			}
		}
	}
	centerPanel := StylePanel.Width(centerW).Height(mainHeight).Render(centerContent.String())

	// ========== RIGHT PANEL: Stats ==========
	rightContent := new(strings.Builder)
	rightContent.WriteString(StyleStatLabel.Render(" STATS ") + "\n")
	rightContent.WriteString(strings.Repeat("─", rightW-2) + "\n")
	rightContent.WriteString("\n")

	// Files
	rightContent.WriteString(StyleStatLabel.Render(" Files    "))
	rightContent.WriteString(StyleStatValue.Render(fmt.Sprintf("%d", m.TotalFiles)) + "\n")

	// Errors
	rightContent.WriteString(StyleStatLabel.Render(" Errors   "))
	if m.ErrorCount > 0 {
		rightContent.WriteString(StyleStatError.Render(fmt.Sprintf("%d", m.ErrorCount)) + "\n")
	} else {
		rightContent.WriteString(StyleStatSuccess.Render("0") + "\n")
	}

	// Warnings
	rightContent.WriteString(StyleStatLabel.Render(" Warnings "))
	if m.WarnCount > 0 {
		rightContent.WriteString(StyleStatWarning.Render(fmt.Sprintf("%d", m.WarnCount)) + "\n")
	} else {
		rightContent.WriteString(StyleStatSuccess.Render("0") + "\n")
	}

	rightContent.WriteString("\n")
	rightContent.WriteString(StyleFooterHint.Render(" Files watched "))
	rightContent.WriteString("in this session\n")
	rightPanel := StylePanel.Width(rightW).Height(mainHeight).Render(rightContent.String())

	// ========== DETAIL PANEL ==========
	detailView := ""
	if m.ShowDetail && len(m.Violations) > 0 && m.Cursor < len(m.Violations) {
		v := m.Violations[m.Cursor]
		sevLabel := StyleViolationError.Render(" ERROR ")
		if v.Severity == spec.SeverityWarning {
			sevLabel = StyleViolationWarning.Render(" WARNING ")
		}

		detailContent := fmt.Sprintf(
			"%s %s  %s\n\n%s %s\n%s %d\n%s %s\n%s\n%s %s",
			StyleDetailKey.Render("File:"),
			StyleDetailValue.Render(v.File),
			sevLabel,
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
		detailView = StyleDetail.Width(m.Width - 2).Height(detailHeight - 1).Render(detailContent)
	}

	// ========== FOOTER ==========
	footerContent := fmt.Sprintf(
		" %s %s %s %s %s %s %s %s %s",
		StyleFooterKey.Render("↑↓"),
		StyleFooterHint.Render("Navigate"),
		StyleFooterKey.Render("ENTER"),
		StyleFooterHint.Render("Details"),
		StyleFooterKey.Render("C"),
		StyleFooterHint.Render("Clear"),
		StyleFooterKey.Render("Q"),
		StyleFooterHint.Render("Quit"),
		StyleFooterLatency.Render(fmt.Sprintf(" ⚡ %s ", m.Latency)),
	)
	footer := StyleFooter.Width(m.Width).Render(footerContent)

	// ========== COMBINE ==========
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)

	result := StyleBase.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			panels,
			detailView,
			footer,
		),
	)

	return result
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
