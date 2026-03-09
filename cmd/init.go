package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a spec.md file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat("spec.md"); err == nil {
			fmt.Println("spec.md already exists. Remove it first or modify it manually.")
			os.Exit(1)
		}

		tmpl, err := template.New("spec").Parse(specTemplate)
		if err != nil {
			return err
		}

		file, err := os.Create("spec.md")
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		err = tmpl.Execute(file, nil)
		if err != nil {
			return err
		}

		fmt.Println("Created spec.md")
		return nil
	},
}

const specTemplate = `# specwatch

## stack
- language: typescript
- framework: next.js@14
- styling: tailwind
- runtime: node@20

## structure
- components: src/components/**
- api routes: src/app/api/**
- utilities: src/lib/**
- types: src/types/**
- tests: **/*.test.ts, **/*.spec.ts

## naming
- components: PascalCase
- functions: camelCase
- files: kebab-case
- constants: SCREAMING_SNAKE_CASE
- interfaces: PascalCase prefixed with I

## forbidden
- pattern: "console.log"
  message: use logger utility from @/lib/logger
- pattern: "any"
  message: no any types — use unknown or explicit type
- pattern: "style={{"
  message: no inline styles — use tailwind classes
- import: "lodash"
  message: use native ES methods
- import: "moment"
  message: use date-fns instead

## required
- async functions: try/catch
- api routes: return type { data, error }
- components: must have displayName
- new files in src/components: must have matching *.test.ts

## architecture
- no direct db calls outside src/lib/db
- no business logic in components — belongs in hooks or lib
- server components by default — client components need explicit justification

## limits
- max file lines: 300
- max function lines: 50
- max imports per file: 20
- max component props: 8
`
