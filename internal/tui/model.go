package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
)

type ActivityItem struct {
	File      string
	Clean     bool
	Timestamp time.Time
}

type Model struct {
	Activity   []ActivityItem
	Violations []analyzer.Violation
	TotalFiles int
	ErrorCount int
	WarnCount  int

	Cursor      int
	Width       int
	Height      int
	Analyzing   bool
	LastResults string
}

type NewViolationMsg struct {
	File       string
	Violations []analyzer.Violation
	Duration   time.Duration
}

func InitialModel() Model {
	return Model{
		Activity:   []ActivityItem{},
		Violations: []analyzer.Violation{},
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
			m.Violations = []analyzer.Violation{}
			m.ErrorCount = 0
			m.WarnCount = 0
			m.Cursor = 0
		}

	case NewViolationMsg:
		// Update Activity
		m.Activity = append([]ActivityItem{{
			File:      msg.File,
			Clean:     len(msg.Violations) == 0,
			Timestamp: time.Now(),
		}}, m.Activity...)
		if len(m.Activity) > 20 {
			m.Activity = m.Activity[:20]
		}

		// Update Violations (remove old ones for this file, add new ones)
		var newViolations []analyzer.Violation
		for _, v := range m.Violations {
			if v.File != msg.File {
				newViolations = append(newViolations, v)
			}
		}
		newViolations = append(newViolations, msg.Violations...)
		m.Violations = newViolations

		// Update Stats
		m.ErrorCount = 0
		m.WarnCount = 0
		uniqueFiles := make(map[string]bool)
		for _, v := range m.Violations {
			if strings.Contains(string(v.Severity), "error") {
				m.ErrorCount++
			} else {
				m.WarnCount++
			}
			uniqueFiles[v.File] = true
		}

		// This is a bit simplistic, we'd want a more robust way to track TotalFiles
		// For now let's just count unique files with violations + clean files we've seen
		seenFiles := make(map[string]bool)
		for _, a := range m.Activity {
			seenFiles[a.File] = true
		}
		m.TotalFiles = len(seenFiles)

		m.LastResults = msg.Duration.String()
	}

	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder

	// Title bar
	title := fmt.Sprintf(" specwatch v0.1.0 %s ", getStatusIndicator(m.Analyzing))
	sb.WriteString(StyleTitle.Render(title))
	sb.WriteString("\n\n")

	// Activity column
	sb.WriteString(StyleHeader.Render("ACTIVITY"))
	sb.WriteString("\n")
	for i, item := range m.Activity {
		marker := "○ "
		style := StyleFileClean
		if !item.Clean {
			marker = "● "
			style = StyleFileViolation
		}
		cursor := " "
		if i == m.Cursor && len(m.Violations) > 0 {
			cursor = ">"
		}
		filename := filepath.Base(item.File)
		age := time.Since(item.Timestamp).Round(time.Second)
		sb.WriteString(fmt.Sprintf("%s%s%s %s\n", cursor, style.Render(marker), filename, StyleTime.Render(age.String())))
	}

	// Violations column
	sb.WriteString("\n" + StyleHeader.Render("VIOLATIONS"))
	sb.WriteString("\n")
	if len(m.Violations) == 0 {
		sb.WriteString(StyleTime.Render("No violations"))
	} else {
		for i, v := range m.Violations {
			var cursor string
			if i == m.Cursor {
				cursor = StyleViolationItem.Render("✗ ")
			} else {
				cursor = "  "
			}
			filename := filepath.Base(v.File)
			ruleName := strings.Split(v.Rule, ".")[0]
			sb.WriteString(fmt.Sprintf("%s%s:%d %s\n", cursor, filename, v.Line, StyleRule.Render(ruleName)))
		}
	}

	// Stats column
	sb.WriteString("\n" + StyleHeader.Render("STATS"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %d files\n", StyleFileClean.Render("✓"), m.TotalFiles))
	sb.WriteString(fmt.Sprintf("%s %d errors\n", StyleFileViolation.Render("✗"), m.ErrorCount))
	sb.WriteString(fmt.Sprintf("%s %d warnings\n", StyleTime.Render("⚠"), m.WarnCount))

	// Detail panel at bottom
	if m.Cursor < len(m.Violations) {
		v := m.Violations[m.Cursor]
		sb.WriteString("\n")
		sb.WriteString(StyleDetail.Render(
			fmt.Sprintf("%s %s:%d [%s]", StyleFileViolation.Render("✗"), filepath.Base(v.File), v.Line, v.Rule),
		))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Found:    %s\n", v.Excerpt))
		sb.WriteString(fmt.Sprintf("Fix:      %s\n", v.Suggestion))
	}

	// Footer with timing
	sb.WriteString("\n")
	sb.WriteString(StyleTime.Render(m.LastResults + " "))

	return sb.String()
}

func getStatusIndicator(analyzing bool) string {
	if analyzing {
		return "● watching"
	}
	return "○ idle"
}
