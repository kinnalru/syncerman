# Go Code Style Guidelines

This guide provides comprehensive coding standards for Go development in this project, combining project-specific guidelines with Golang best practices.

## Imports

- Use standard library imports first, then third-party, then internal
- Group imports with blank lines between groups
- Use `goimports` to automatically manage imports
- Prefer named imports for clarity when packages have similar names
- Avoid renaming imports except to avoid name collisions
- Prefer to rename the most local or project-specific import when collisions occur
- Import only packages that are actually used

## Formatting

- Always run `go fmt` before committing
- Use `gofmt` with default settings
- Use tabs for indentation (Go standard)
- Maximum line length: 120 characters
- Avoid uncomfortably long lines, but don't add artificial line breaks
- Break lines based on semantics, not length
- Place opening braces on the same line as the statement
- No semicolons except for `for` loop clauses

## Naming Conventions

### Packages
- Lowercase, single words when possible
- No underscores or mixedCaps
- Short, concise, evocative names
- Avoid meaningless names like `util`, `common`, `misc`, `api`, `types`, `interfaces`
- Omit package name from exported identifiers (e.g., `chubby.File` not `chubby.ChubbyFile`)

### Exports
- PascalCase (e.g., `Hello`, `CalculateTotal`)

### Private
- camelCase (e.g., `hello`, `calculateTotal`)

### Constants
- PascalCase for exported, camelCase for unexported

### Interfaces
- Usually -er suffix (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces named after the method plus -er suffix
- Honor canonical method names: `Read`, `Write`, `Close`, `Flush`, `String`

### Errors
- Should have "Error" suffix (e.g., `ValidationError`)

### Initialisms and Acronyms
- Preserve case: `URL` not `Url`, `ID` not `Id`
- ServeHTTP not ServeHttp, appID not appId

### Variables
- Short names for local variables with limited scope
- Prefer `c` to `lineCount`, `i` to `sliceIndex`
- The further from declaration, the more descriptive the name
- One or two letters sufficient for method receivers
- Avoid generic names like `me`, `this`, `self`

## Error Handling

- Always check errors, never ignore them
- Use `errors.Is()` and `errors.As()` for error comparisons
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Prefer `"context: %w"` format for wrapping
- Error strings should not be capitalized (unless proper nouns) or end with punctuation
- Handle errors at the appropriate level
- Return errors as the last return value
- Use `_, ok` pattern to distinguish missing from zero values
- Indent error flow, keep normal path at minimal indentation

## Types

- Prefer concrete types over interfaces when possible
- Define interfaces where they are used (accept interfaces, return structs)
- Interfaces belong in the package that uses values, not the package that implements them
- Don't define interfaces for mocking; use concrete types
- Don't define interfaces before they are used
- Use `type` for domain-specific types that add clarity
- Design types so zero values are useful

### Value vs Pointer Receivers
- Value: for immutable operations, small structs, basic types
- Pointer: for mutable operations, large structs, structs with sync.Mutex
- If receiver contains mutex or similar, must use pointer
- If method mutates receiver, must use pointer
- Don't mix receiver types; choose one consistently
- When in doubt, use pointer receiver

## Functions

- Keep functions small and focused
- Use multiple return values, error is last
- Limit parameters to 3-4, consider struct for more
- Document exported functions with examples on non-obvious use cases
- Prefer synchronous functions over asynchronous ones
- Use `defer` for cleanup (close files, unlock mutexes)
- Don't use panic for normal error handling
- Named result parameters can clarify intent, but don't use them just to enable naked returns


## Structs and Fields

- Exported fields: PascalCase
- Unexported fields: camelCase
- Group related fields together
- Document exported structs and fields
- Be careful copying structs from other packages (may cause aliasing)
- Don't copy values if methods are on pointer type

## Context

- Most functions that use `Context` should accept it as first parameter
- Pass Context explicitly along call chain from incoming requests
- Default to passing Context; use `context.Background()` only with good reason
- Don't add Context member to struct type
- Don't create custom Context types
- Use Context for cancellation, deadlines, tracing, credentials

```go
func Process(ctx context.Context, data Data) (Result, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
}
```

## Comments

- Exported declarations must have comments
- Comments should be complete sentences
- Use `//` for documentation, not `/* */`
- Write godoc comments for all exported code
- Package comments should be in doc.go or at package declaration
- Comments should begin with the name being described and end in period
- Package comments must be adjacent to package clause

## Project Structure

- `./main.go`: Application entry point
- `internal/`: Private application code (not importable externally)
- `pkg/`: Public libraries (if any)
- Test files alongside source with `_test.go` suffix

## Best Practices

- Use `const` for constant values, avoid magic numbers
- Implement `String()` method for custom types when beneficial for string representation
- Use `sync.Pool` for object pooling when appropriate
- Prefer channels over mutexes for coordination when possible
- Prefer synchronous functions to keep goroutines localized
- Goroutine lifetimes should be obvious; leaks cause problems
- Don't pass pointers just to save bytes
- Don't use `*string` or `*interface{}` as parameters
- Declare empty slices as `var t []string` not `t := []string{}`
- Use `make` for maps, slices, channels; `new` for other allocations
- Range loops: use blank identifier (`_`) to ignore unwanted values

## Concurrency

- Keep concurrent code simple enough that goroutine lifetimes are obvious
- Document when and why goroutines exit
- Avoid blocking on channels that may block indefinitely
- Prefer synchronous APIs over callbacks for easier testing
- Use `defer` with mutex unlocks
- Be careful with shared mutable state

## Testing

- Table-driven tests for multiple scenarios
- Use `testing` package or testify/assert
- Test both happy path and error cases
- Use `t.Run()` for subtests
- Mock external dependencies using interfaces
- Keep tests independent and deterministic
- Fail with helpful messages: actual != expected
- Include examples for packages
- Use `import .` only in test files when necessary

## Security

- Use `crypto/rand` for keys, not `math/rand`
- Never log secrets or keys
- Redact sensitive data from error messages
- Use `strconv.Quote` or `fmt.Sprintf("%q")` for user input in logs

## Tools

Run these before committing:
```bash
go fmt ./...
go vet ./...
goimports -w .
golangci-lint run
```

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Style Guide (Google)](https://google.github.io/styleguide/go/decisions)

