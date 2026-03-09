package analyzer

import "github.com/rajeshshrirao/specwatch/internal/spec"

type Violation struct {
	File       string        `json:"file"`
	Line       int           `json:"line"`
	Rule       string        `json:"rule"`
	Severity   spec.Severity `json:"severity"`
	Excerpt    string        `json:"excerpt"`
	Suggestion string        `json:"suggestion"`
}
