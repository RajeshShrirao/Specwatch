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
	Activity    []ActivityItem
	Violations  []InternalViolation
	TotalFiles  int
	ErrorCount  int
	WarnCount   int
	Latency     string
	ShowDetail  bool
	Cursor      int
	Width       int
	Height      int
	Analyzing   bool
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
			m.WarnCount = 1 // Just kidding, 0
			m.WarnCount = 0
			m.Cursor = 0
		case "enter":
			if len(m.Violations) > 0 {
				m.ShowDetail = !m.ShowDetail
			}
		}

	case NewViolationMsg:
		now := time.Now()
		// Update Activity
		m.Activity = append([]ActivityItem{{
			File:      msg.File,
			Clean:     len(msg.Violations) == 0,
			Timestamp: now,
		}}, m.Activity...)
		if len(m.Activity) > 50 {
			m.Activity = m.Activity[:50]
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
		
		// Sort: Severity (Error > Warning), then Recency (Newest first)
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
		return "Initializing..."
	}

	// Calculate dimensions
	footerHeight := 1
	detailHeight := 0
	if m.ShowDetail {
		detailHeight = 8 // Fixed height for detail bar
	}
	
	mainHeight := m.Height - footerHeight - detailHeight - 2 // -2 for title/spacing?
	if mainHeight < 0 { mainHeight = 0 }

	leftWidth := int(float64(m.Width) * 0.3)
	rightWidth := int(float64(m.Width) * 0.2)
	centerWidth := m.Width - leftWidth - rightWidth

	// Styles with dynamic widths
	leftStyle := StylePanel.Copy().Width(leftWidth - 2).Height(mainHeight)
	centerStyle := StylePanelActive.Copy().Width(centerWidth - 2).Height(mainHeight)
	rightStyle := StylePanel.Copy().Width(rightWidth - 2).Height(mainHeight)

	// --- Left Panel: Activity Feed ---
	var activityLines []string
	activityLines = append(activityLines, StyleTitle.Render("ACTIVITY"))
	
	// Slice activity to fit mainHeight
	maxActivity := mainHeight - 2
	visibleActivity := m.Activity
	if len(visibleActivity) > maxActivity {
		visibleActivity = visibleActivity[:maxActivity]
	}

	for _, item := range visibleActivity {
		dot := StyleSuccess.Render("●")
		if !item.Clean {
			dot = StyleError.Render("●")
		}
		filename := filepath.Base(item.File)
		elapsed := time.Since(item.Timestamp).Round(time.Second).String()
		line := fmt.Sprintf("%s %s %s", dot, filename, StyleMuted.Render(elapsed))
		activityLines = append(activityLines, line)
	}
	leftView := leftStyle.Render(strings.Join(activityLines, "\n"))

	// --- Center Panel: Violations ---
	var violationLines []string
	violationLines = append(violationLines, StyleTitle.Render("VIOLATIONS"))
	if len(m.Violations) == 0 {
		violationLines = append(violationLines, StyleMuted.Render("No violations detected"))
	} else {
		// Calculate viewport for violations
		maxViolations := mainHeight - 2
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
			prefix := "  "
			style := StyleViolationItem
			if i == m.Cursor {
				prefix = StyleAccent.Render("→ ")
				style = StyleViolationSelected.Copy().Width(centerWidth - 6)
			}
			
			loc := fmt.Sprintf("%s:%d", filepath.Base(v.File), v.Line)
			rule := v.Rule
			if len(rule) > 20 { rule = rule[:17] + "..." }
			
			line := style.Render(fmt.Sprintf("%s%s %s", prefix, loc, StyleMuted.Render(rule)))
			violationLines = append(violationLines, line)
		}
	}
	centerView := centerStyle.Render(strings.Join(violationLines, "\n"))

	// --- Right Panel: Stats ---
	var statLines []string
	statLines = append(statLines, StyleTitle.Render("STATS"))
	statLines = append(statLines, "", "Files watched")
	statLines = append(statLines, StyleStatValue.Render(fmt.Sprintf("%d", m.TotalFiles)), "")
	statLines = append(statLines, "Errors")
	statLines = append(statLines, StyleError.Copy().Inherit(StyleStatValue).Render(fmt.Sprintf("%d", m.ErrorCount)), "")
	statLines = append(statLines, "Warnings")
	statLines = append(statLines, StyleWarning.Copy().Inherit(StyleStatValue).Render(fmt.Sprintf("%d", m.WarnCount)))
	rightView := rightStyle.Render(strings.Join(statLines, "\n"))

	// --- Bottom Detail Bar ---
	detailView := ""
	if m.ShowDetail && len(m.Violations) > 0 {
		v := m.Violations[m.Cursor]
		detailStyle := StylePanel.Copy().Width(m.Width - 2).Height(detailHeight - 2)
		
		content := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
			StyleDetailKey.Render("File"), StyleDetailValue.Render(v.File),
			StyleDetailKey.Render("Line"), StyleDetailValue.Render(fmt.Sprintf("%d", v.Line)),
			StyleDetailKey.Render("Category"), StyleDetailValue.Render(v.Rule),
			StyleDetailKey.Render("Snippet"), StyleDetailValue.Render(v.Excerpt),
			StyleDetailKey.Render("Fix"), StyleAccent.Render(v.Suggestion),
		)
		detailView = detailStyle.Render(content)
	}

	// --- Footer ---
	footerLeft := fmt.Sprintf(
		"%s %s %s %s %s %s %s %s",
		StyleFooterKey.Render("[j/k]"), StyleFooterHint.Render("Navigate"),
		StyleFooterKey.Render("[enter]"), StyleFooterHint.Render("Details"),
		StyleFooterKey.Render("[c]"), StyleFooterHint.Render("Clear"),
		StyleFooterKey.Render("[q]"), StyleFooterHint.Render("Quit"),
	)
	footerRight := StyleLatency.Render(m.Latency)
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom,
		footerLeft,
		strings.Repeat(" ", max(0, m.Width-lipgloss.Width(footerLeft)-lipgloss.Width(footerRight))),
		footerRight,
	)

	// Combine everything
	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftView, centerView, rightView)
	
	finalView := lipgloss.JoinVertical(lipgloss.Left,
		mainPanels,
		detailView,
		footer,
	)

	return StyleMain.Render(finalView)
}

func max(a, b int) int {
	if a > b { return a }
	return b
}
