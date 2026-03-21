# specwatch

## stack
- language: go
- framework:
- runtime: go@1.22

## structure
- commands: cmd/**
- internal packages: internal/**
- main packages: pkg/**
- tests: **/*_test.go

## naming
- packages: lowercase, short names
- functions: PascalCase (exported), camelCase (unexported)
- files: snake_case.go
- constants: PascalCase
- interfaces: PascalCase with -er suffix when appropriate
- errors: error types prefixed with Err

## forbidden
- pattern: "fmt.Printf"
  message: use structured logging (zap, zerolog, or klog)
- pattern: "panic("
  message: avoid panics — return errors instead
- pattern: "//TODO"
  message: use //FIXME or track in issue tracker
- pattern: "time.Sleep"
  message: use context with timeout for operations
- import: "encoding/json"
  message: use json-iterator/go for better performance

## required
- exported functions: must have doc comments
- errors: wrap with fmt.Errorf("...: %w", err)
- main: must use cobra for CLI structure
- config: use viper for configuration management

## architecture
- no business logic in cmd/ — belongs in internal/
- no circular dependencies between internal packages
- use interfaces for external dependencies
- pass context as first argument to functions

## limits
- max file lines: 500
- max function lines: 50
- max imports per file: 20
- max parameters per function: 6
