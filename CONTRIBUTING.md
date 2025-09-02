# Contributing to RuleStack (RFH)

Welcome! We're excited you're interested in contributing to RuleStack. This document provides guidelines to help you contribute effectively.

## Philosophy

**We prioritize working software over perfect code.** If it works and improves the project, we want it!

## Getting Started

### Development Environment

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd rfh
   ```

2. **Start development services**
   ```bash
   docker-compose up -d
   ```

3. **Build the CLI**
   ```bash
   go build -o dist/rfh ./cmd/cli
   ```

4. **Run tests**
   ```bash
   # Run all tests
   ./run-tests.ps1  # Windows
   # ./run-tests.sh   # Linux/Mac
   
   # Run specific test suites
   go test ./...
   cd cucumber-testing && npm test
   ```

For detailed setup instructions, see [Development Setup](docs/development/setup.md).

## How to Contribute

### 1. Issues and Bug Reports

- Search existing issues before creating new ones
- Use the issue templates when available
- Include system information (OS, RFH version, Go version)
- Provide minimal reproduction steps

### 2. Code Contributions

#### Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** following our coding standards
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Ensure all tests pass** locally
6. **Submit a pull request** with a clear description

#### Coding Standards

- **Follow Go conventions** - Use `gofmt`, `golint`, and `go vet`
- **Write tests** - All new features should have corresponding tests
- **Document public APIs** - Include comments for exported functions
- **Keep commits focused** - One logical change per commit
- **Use descriptive commit messages** - Follow conventional commit format

#### Project-Specific Guidelines

**IMPORTANT**: Follow all rules in the `.rulestack/project/` directory:

- **No backward compatibility** - Remove deprecated code immediately
- **Simple implementation** - Choose the simplest solution that works
- **Save implementation plans** - Document major changes in `planning/` folder

### 3. Documentation

- Update relevant documentation when making changes
- Keep README.md current with new features
- Add examples for new CLI commands
- Update troubleshooting guides for common issues

### 4. Testing

We use multiple testing approaches:

- **Unit tests** - Go tests in `internal/` packages
- **Integration tests** - Cucumber BDD tests in `cucumber-testing/`
- **End-to-end tests** - Full workflow testing

#### Running Tests

```bash
# Unit tests
go test ./...

# Cucumber tests
cd cucumber-testing
npm install
npm test

# Specific feature
npx cucumber-js features/auth.feature
```

## Development Workflow

### Setting Up for Development

1. **Initialize development environment**
   ```bash
   # Start services
   docker-compose up -d
   
   # Verify setup
   go test ./internal/config
   ```

2. **Test CLI locally**
   ```bash
   # Build
   go build -o dist/rfh ./cmd/cli
   
   # Test basic functionality
   ./dist/rfh --version
   ./dist/rfh init
   ```

### Making Changes

1. **Create feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Implement changes**
   - Follow coding standards
   - Add tests for new functionality
   - Update documentation

3. **Test thoroughly**
   ```bash
   # Run full test suite
   ./run-tests.ps1
   
   # Test specific areas
   go test ./internal/cli
   npx cucumber-js features/pack.feature
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   git push origin feature/my-feature
   ```

## Areas We Need Help With

### High Priority

- **Windows testing** - Ensure cross-platform compatibility
- **Security validation** - Improve package scanning and validation
- **Documentation** - Keep docs current with rapid development
- **Error handling** - Improve user experience with better error messages

### Medium Priority

- **Performance optimization** - Package operations and network requests
- **CI/CD improvements** - GitHub Actions workflow enhancements
- **Registry features** - Private registry improvements
- **Editor integrations** - Support for additional AI editors

### Good First Issues

- **Documentation fixes** - Typos, missing examples, outdated info
- **Test coverage** - Add tests for uncovered code paths
- **Error message improvements** - Make errors more user-friendly
- **CLI help text** - Improve command descriptions and examples

## Code Review Process

### For Contributors

- **Keep PRs focused** - One feature or fix per PR
- **Include tests** - New code should have corresponding tests
- **Update docs** - Include documentation changes
- **Be responsive** - Address review feedback promptly

### For Reviewers

- **Be constructive** - Focus on code quality and functionality
- **Test the changes** - Pull and test the PR locally
- **Check documentation** - Ensure docs are updated
- **Approve when ready** - Don't let perfect be the enemy of good

## Release Process

We follow semantic versioning (semver):

- **Major** (`1.0.0`) - Breaking changes
- **Minor** (`0.1.0`) - New features, backward compatible
- **Patch** (`0.0.1`) - Bug fixes, backward compatible

## Security

### Reporting Security Issues

Please report security vulnerabilities privately by emailing the maintainers. Do not create public issues for security problems.

### Security Guidelines

- **Validate all inputs** - Sanitize user data and file paths
- **Never commit secrets** - Use environment variables for sensitive data
- **Follow least privilege** - Minimal required permissions
- **Audit dependencies** - Keep dependencies updated

## Community Guidelines

### Code of Conduct

- **Be respectful** - Treat all contributors with respect
- **Be inclusive** - Welcome developers of all skill levels
- **Be collaborative** - Work together toward common goals
- **Be constructive** - Provide helpful feedback

### Communication

- **GitHub Issues** - Bug reports, feature requests
- **GitHub Discussions** - General questions, ideas
- **Pull Requests** - Code contributions, reviews

## Questions?

If you have questions about contributing:

1. Check the [documentation](docs/)
2. Search existing [issues](https://github.com/your-org/rfh/issues)
3. Create a new issue or discussion
4. Reach out to maintainers

---

**Thank you for contributing to RuleStack!** 🚀

Your contributions help make AI rulesets accessible and shareable for everyone.