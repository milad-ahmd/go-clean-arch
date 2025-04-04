# Contributing to Go Clean Architecture

Thank you for considering contributing to this project! Here are some guidelines to help you get started.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct.

## How Can I Contribute?

### Reporting Bugs

- Check if the bug has already been reported in the Issues section
- Use the bug report template when creating a new issue
- Include detailed steps to reproduce the bug
- Include any relevant logs or error messages

### Suggesting Enhancements

- Check if the enhancement has already been suggested in the Issues section
- Use the feature request template when creating a new issue
- Describe the feature in detail and why it would be valuable

### Pull Requests

1. Fork the repository
2. Create a new branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests to ensure they pass
5. Commit your changes using conventional commit messages
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Development Setup

1. Clone the repository
2. Install dependencies with `go mod tidy`
3. Copy `.env.example` to `.env` and configure as needed
4. Run the application with `go run cmd/api/main.go`

## Coding Standards

- Follow Go best practices and style guidelines
- Write tests for new features
- Document public functions and types
- Use meaningful variable and function names

## Commit Message Guidelines

We use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for adding or modifying tests
- `chore:` for maintenance tasks

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
