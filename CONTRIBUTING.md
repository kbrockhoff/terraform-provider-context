# Contributing to Terraform Provider Context

Thank you for your interest in contributing to the Terraform Provider Context! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Contributions](#making-contributions)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project follows a standard Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [kbrockhoff@gmail.com](mailto:kbrockhoff@gmail.com).

### Our Standards

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

## Getting Started

### Prerequisites

- **Go**: Version 1.21 or later
- **Terraform**: Latest version for testing
- **Git**: For version control
- **Make**: For running development tasks

### Development Tools (Optional but Recommended)

- **golangci-lint**: For code linting
- **tfplugindocs**: For documentation generation
- **goreleaser**: For release builds

## Development Setup

1. **Fork and Clone the Repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/terraform-provider-context.git
   cd terraform-provider-context
   ```

2. **Set Up Development Environment**
   ```bash
   make dev-setup
   ```

3. **Install Provider Locally**
   ```bash
   make install
   ```

4. **Run Tests**
   ```bash
   make test
   ```

5. **Verify Everything Works**
   ```bash
   make check
   ```

## Making Contributions

### Types of Contributions

We welcome several types of contributions:

- **🐛 Bug Reports**: Help us identify and fix issues
- **✨ Feature Requests**: Propose new functionality
- **📚 Documentation**: Improve or add documentation
- **🔧 Code Contributions**: Bug fixes, features, refactoring
- **🧪 Tests**: Add or improve test coverage
- **📦 Examples**: Add usage examples

### Before You Start

1. **Check Existing Issues**: Search for existing issues or discussions
2. **Create an Issue**: For significant changes, create an issue first to discuss
3. **Read the Design**: Understand the project architecture in `docs/design.md`
4. **Review Recent PRs**: See what others are working on

## Development Workflow

### Project Structure

```
terraform-provider-context/
├── internal/
│   ├── provider/       # Provider implementation
│   ├── datasource/     # Data source implementations  
│   ├── core/           # Business logic
│   └── models/         # Data models
├── examples/           # Usage examples
├── docs/               # Documentation
└── tests/              # Integration tests
```

### Core Components

- **Provider**: Minimal configuration (cloud_provider, tag_prefix)
- **Data Source**: Main functionality and all configuration options
- **Core Logic**: Business rules for naming, tagging, validation
- **Cloud Providers**: Platform-specific tag formatting

### Development Tasks

| Command | Description |
|---------|-------------|
| `make build` | Build the provider binary |
| `make test` | Run unit tests |
| `make lint` | Run linting checks |
| `make install` | Install provider locally |
| `make test-examples` | Validate example configurations |
| `make clean` | Clean build artifacts |

## Testing

### Test Types

1. **Unit Tests** (`internal/core/*_test.go`)
   - Test business logic in isolation
   - Fast execution, no external dependencies
   - Required for all new functionality

2. **Integration Tests** 
   - Test provider with Terraform
   - Use acceptance testing framework
   - Run with `TF_ACC=1 go test`

3. **Example Tests**
   - Validate example configurations
   - Run with `make test-examples`

### Writing Tests

#### Unit Test Example

```go
func TestNameGenerator_Generate(t *testing.T) {
    tests := []struct {
        name        string
        namespace   string
        resourceName string
        environment string
        want        string
        wantErr     bool
    }{
        {
            name:         "standard format",
            namespace:    "myorg",
            resourceName: "app", 
            environment:  "prod",
            want:         "myorg-app-prod",
            wantErr:      false,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ng := &NameGenerator{
                Namespace:   tt.namespace,
                Name:        tt.resourceName,
                Environment: tt.environment,
            }
            got, err := ng.Generate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Generate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Test Coverage

- Aim for >90% test coverage for new code
- Include edge cases and error conditions
- Test all cloud provider variations
- Verify input validation

### Testing Guidelines

- **Test file naming**: `*_test.go`
- **Test function naming**: `TestFunctionName_Scenario`
- **Table-driven tests**: Use for multiple test cases
- **Error testing**: Verify both success and failure cases
- **Mock external dependencies**: Use interfaces for testability

## Code Style

### Go Style Guidelines

- Follow standard Go conventions
- Use `gofmt` and `gofumpt` for formatting
- Pass `golangci-lint` checks
- Write clear, self-documenting code
- Use meaningful variable and function names

### Code Organization

- **Single responsibility**: Each function should do one thing well
- **Interface segregation**: Keep interfaces small and focused
- **Dependency injection**: Use interfaces for testability
- **Error handling**: Always handle errors appropriately

### Documentation

- **Godoc comments**: All public functions and types
- **README updates**: Update for new features
- **Examples**: Provide usage examples
- **Design docs**: Update for architectural changes

### Validation Rules

- Input validation at data source level
- Business logic validation in core package
- Clear error messages for users
- Consistent error formats

## Submitting Changes

### Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Write code following style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Test Changes**
   ```bash
   make pre-commit
   ```

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `ci`: CI/CD changes
- `chore`: Maintenance tasks

#### Examples

```bash
feat: add support for custom tag prefixes
fix: resolve issue with GCP tag formatting
docs: update README with new examples
test: add unit tests for validation logic
```

### Pull Request Guidelines

- **Fill out the PR template completely**
- **Keep PRs focused and atomic**
- **Include tests for new functionality**
- **Update documentation as needed**
- **Ensure CI passes**
- **Request review from maintainers**

### Review Process

1. **Automated Checks**: CI must pass
2. **Code Review**: Maintainer review required
3. **Testing**: Manual testing may be requested
4. **Documentation**: Verify docs are updated
5. **Merge**: Squash and merge after approval

## Architecture Guidelines

### Provider Design

- **Minimal Provider Config**: Only cloud_provider and tag_prefix
- **Rich Data Source**: All other configuration in data source
- **Cloud Abstraction**: Support multiple cloud providers
- **Stateless Operation**: No state maintained between calls

### Adding New Features

#### New Configuration Field

1. Add to `DataSourceConfig` in `core/tags.go`
2. Add to schema in `datasource/context.go`
3. Add validation in `core/validation.go`
4. Add tests for all new functionality
5. Update examples and documentation

#### New Cloud Provider

1. Implement `CloudProvider` interface in `core/cloud.go`
2. Add provider to `GetCloudProvider` function
3. Add validation to `ValidCloudProviders`
4. Add comprehensive tests
5. Update documentation

#### New Tag Category

1. Add processing logic to `TagProcessor.Process()`
2. Add configuration fields if needed
3. Consider data tags vs regular tags
4. Add feature toggle if optional
5. Test with all cloud providers

### Performance Considerations

- **Caching**: Cache expensive operations (git info)
- **Memory**: Minimize allocations in hot paths
- **External calls**: Minimize and cache results
- **Validation**: Fail fast for invalid inputs

## Release Process

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. **Update Documentation**
   - README version references
   - CHANGELOG entries
   - Example configurations

2. **Test Thoroughly**
   - All unit tests pass
   - Integration tests pass
   - Example validation

3. **Tag Release**
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

4. **GitHub Actions** handles the rest:
   - Build cross-platform binaries
   - Create GitHub release
   - Sign artifacts
   - Update documentation

## Getting Help

### Resources

- **Documentation**: Check existing docs first
- **Issues**: Search existing issues
- **Discussions**: Use GitHub Discussions for questions
- **Examples**: Look at usage examples

### Communication

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Email**: [kbrockhoff@gmail.com](mailto:kbrockhoff@gmail.com) for sensitive issues

### Response Times

- **Issues**: We aim to respond within 2-3 business days
- **PRs**: Reviews typically within 1 week
- **Security Issues**: Within 24 hours

## Recognition

Contributors who make significant contributions will be:

- Listed in release notes
- Added to contributor acknowledgments
- Invited to become maintainers (for ongoing contributors)

## Questions?

Don't hesitate to ask questions! We're here to help and want to make contributing as smooth as possible.

---

**Thank you for contributing to Terraform Provider Context!** 🎉