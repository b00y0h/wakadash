# Contributing to wakadash

Thank you for your interest in contributing!

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/wakadash`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes with a descriptive message
7. Push to your fork and open a Pull Request

## Development Requirements

- Go 1.21 or later
- A WakaTime account or self-hosted Wakapi instance for integration testing

## Code Style

- Run `gofmt` before committing
- Run `go vet ./...` and ensure it passes
- Keep functions small and focused
- Add comments for exported types and functions

## Pull Request Guidelines

- Keep PRs focused on a single change
- Include a clear description of what changed and why
- Reference any related issues

## Reporting Issues

Please open an issue on GitHub with:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your Go version (`go version`) and OS

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
