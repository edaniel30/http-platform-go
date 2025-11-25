# Contributing to HTTP Platform Go

Thank you for your interest in contributing to HTTP Platform Go! This document provides guidelines and instructions for contributing.

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/edaniel30/http-platform-go.git
   cd http-platform-go
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

## Code Style

- Follow standard Go conventions and style guidelines
- Use `gofmt` to format your code
- Run `go vet` to check for common mistakes
- Use meaningful variable and function names
- Write clear comments for exported functions and types

## Coding Guidelines

### 1. Comments

- All exported types, functions, and constants must have godoc comments
- Comments should start with the name of the item being documented
- Use complete sentences with proper punctuation

Example:
```go
// Platform is the main HTTP server platform
// It encapsulates server lifecycle, routing, and middleware management
type Platform struct {
    // ...
}
```

### 2. Error Handling

- Always check and handle errors
- Use custom error types when appropriate
- Wrap errors with context using `fmt.Errorf` with `%w`

Example:
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### 3. Configuration

- Use functional options pattern for configuration
- Provide sensible defaults
- Validate configuration in constructors

## Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, well-documented code
   - Add tests for new functionality
   - Update documentation as needed

3. **Run checks**
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add feature: your feature description"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**
   - Provide a clear description of the changes
   - Reference any related issues
   - Ensure all CI checks pass

## Commit Message Guidelines

Use clear and descriptive commit messages:

- **feature**: A new feature
- **hotfix**: A bug fix
- **docs**: Documentation changes

Examples:
```
feat: add support for custom middleware configuration
fix: resolve race condition in platform shutdown
docs: update README with CORS configuration examples
refactor: simplify error handling in config validation
```

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- Operating system and architecture
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant code snippets or logs

## Code Review Process

All contributions will be reviewed by maintainers. We look for:

- Code quality and clarity
- Test coverage
- Documentation completeness
- Adherence to project conventions
- Performance implications

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Reach out to maintainers

Thank you for contributing to HTTP Platform Go! 🚀
