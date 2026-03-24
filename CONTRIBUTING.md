# Contributing to goreadstat

Thank you for your interest in contributing!

## Development Setup

1. Clone the repository
2. Install prerequisites: Go 1.21+, C compiler, zlib, iconv
3. Build the ReadStat library: `make`
4. Run tests: `go test -v ./...`

## Testing

- Add tests for new features
- Ensure all tests pass before submitting PR
- Run `go test -race ./...` to check for race conditions

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Add godoc comments for exported functions

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes with clear commit messages
4. Add tests and documentation
5. Submit a PR with a clear description

## Reporting Issues

- Use GitHub Issues
- Include Go version, OS, and error messages
- Provide minimal reproducible examples
