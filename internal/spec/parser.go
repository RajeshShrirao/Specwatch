package spec

import (
	"os"
	"strconv"
	"strings"

	specerr "github.com/rajeshshrirao/specwatch/internal/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parse reads the spec.md file and returns a RuleSet
func Parse(path string) (*RuleSet, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, specerr.Wrap(err, specerr.ErrCodeIO, "failed to read spec file")
	}

	md := goldmark.New()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	rules := &RuleSet{}

	// Iterate through top-level nodes
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*ast.Heading); ok && heading.Level == 2 {
			title := strings.ToLower(string(heading.Text(content)))

			// Find the list following the heading
			next := heading.NextSibling()
			if next == nil {
				continue
			}

			// If there's a paragraph or other nodes, skip to the list
			for next != nil {
				if _, ok := next.(*ast.List); ok {
					break
				}
				next = next.NextSibling()
			}

			if list, ok := next.(*ast.List); ok {
				parseSection(rules, title, list, content)
			}
		}
	}

	return rules, nil
}

func parseSection(rules *RuleSet, title string, list *ast.List, source []byte) {
	switch title {
	case "stack":
		rules.Stack = parseStack(list, source)
	case "structure":
		rules.Structure = parseStructure(list, source)
	case "naming":
		rules.Naming = parseNaming(list, source)
	case "forbidden":
		rules.Forbidden = parseForbidden(list, source)
	case "required":
		rules.Required = parseRequired(list, source)
	case "architecture":
		rules.Architecture = parseArchitecture(list, source)
	case "limits":
		rules.Limits = parseLimits(list, source)
	}
}

func parseStack(list *ast.List, source []byte) StackRules {
	stack := StackRules{}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "language":
			stack.Language = val
		case "framework":
			stack.Framework = val
		case "styling":
			stack.Styling = val
		case "runtime":
			stack.Runtime = val
		}
	}
	return stack
}

func parseStructure(list *ast.List, source []byte) []StructureRule {
	var rules []StructureRule
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		rules = append(rules, StructureRule{
			Type:    strings.TrimSpace(parts[0]),
			Pattern: strings.TrimSpace(parts[1]),
		})
	}
	return rules
}

func parseNaming(list *ast.List, source []byte) NamingRules {
	naming := NamingRules{}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "components":
			naming.Components = val
		case "functions":
			naming.Functions = val
		case "files":
			naming.Files = val
		case "constants":
			naming.Constants = val
		case "interfaces":
			naming.Interfaces = val
		}
	}
	return naming
}

func parseForbidden(list *ast.List, source []byte) []ForbiddenRule {
	var rules []ForbiddenRule
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		// Get all text from the list item - TextBlock contains the text
		var fullText strings.Builder

		// TextBlock is the container, we need to get text from it
		if textBlock, ok := item.FirstChild().(*ast.TextBlock); ok {
			fullText.WriteString(string(textBlock.Text(source)))
		}

		line := strings.TrimSpace(fullText.String())

		// First split by pattern or import
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "pattern" || key == "import" {
			rule := ForbiddenRule{}

			// Extract pattern/import value and message
			ruleVal := val
			message := ""

			// Look for message: in the value (may not have space before it)
			if idx := strings.Index(val, "message:"); idx != -1 {
				ruleVal = strings.TrimSpace(val[:idx])
				message = strings.TrimSpace(val[idx+8:])
			}

			if key == "pattern" {
				rule.Pattern = strings.Trim(ruleVal, "\"")
			} else {
				rule.Import = strings.Trim(ruleVal, "\"")
			}
			rule.Message = message

			if rule.Pattern != "" || rule.Import != "" {
				rules = append(rules, rule)
			}
		}
	}
	return rules
}

func parseRequired(list *ast.List, source []byte) []RequiredRule {
	var rules []RequiredRule
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		rules = append(rules, RequiredRule{
			Target: strings.TrimSpace(parts[0]),
			Check:  strings.TrimSpace(parts[1]),
		})
	}
	return rules
}

func parseArchitecture(list *ast.List, source []byte) []ArchitectureRule {
	var rules []ArchitectureRule
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		rules = append(rules, ArchitectureRule{
			Description: line,
		})
	}
	return rules
}

func parseLimits(list *ast.List, source []byte) LimitRules {
	limits := LimitRules{}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		line := strings.TrimSpace(string(item.FirstChild().Text(source)))
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val, _ := strconv.Atoi(strings.TrimSpace(parts[1]))

		switch key {
		case "max file lines":
			limits.MaxFileLines = val
		case "max function lines":
			limits.MaxFunctionLines = val
		case "max imports per file":
			limits.MaxImports = val
		case "max component props":
			limits.MaxProps = val
		}
	}
	return limits
}
