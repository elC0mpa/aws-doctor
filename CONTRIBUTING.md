# Contributing to aws-doctor

Thank you for your interest in contributing to aws-doctor! This document provides guidelines and best practices for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Messages](#commit-messages)

## Getting Started

### Prerequisites

- Go 1.21 or later
- AWS credentials configured (for integration testing)
- Git

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork locally:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/aws-doctor.git
   cd aws-doctor
   ```

3. **Add the upstream remote:**
   ```bash
   git remote add upstream https://github.com/elC0mpa/aws-doctor.git
   ```

4. **Verify your remotes:**
   ```bash
   git remote -v
   # Should show:
   # origin    https://github.com/YOUR_USERNAME/aws-doctor.git (fetch)
   # origin    https://github.com/YOUR_USERNAME/aws-doctor.git (push)
   # upstream  https://github.com/elC0mpa/aws-doctor.git (fetch)
   # upstream  https://github.com/elC0mpa/aws-doctor.git (push)
   ```

5. **Install dependencies:**
   ```bash
   go mod download
   ```

## Development Workflow

### Branch Strategy

- **`main`** - Production-ready code, releases are tagged here
- **`development`** - Integration branch for features, **all PRs should target this branch**

### Creating a Feature Branch

Always branch from `development`:

```bash
# Fetch latest changes
git fetch upstream

# Create your feature branch from upstream/development
git checkout -b feat/your-feature-name upstream/development
```

### Keeping Your Branch Updated

Before submitting a PR and when requested by maintainers, rebase your branch against `upstream/development`:

```bash
# Fetch latest changes
git fetch upstream

# Rebase your branch
git checkout your-branch-name
git rebase upstream/development

# Force push if you've already pushed (be careful!)
git push origin your-branch-name --force
```

**Note:** Some contributors may name the original repository remote differently (e.g., `origin` instead of `upstream`). Adjust commands accordingly based on your setup.

### Local Integration Branch (Optional)

If you're working on multiple features and want to test them together locally before they're merged upstream, you can maintain a local integration branch:

```bash
# Create a local-only integration branch based on upstream/development
git checkout -b local/integration upstream/development

# Merge feature branches you want to test together
git merge feat/your-feature-1 --no-edit
git merge feat/your-feature-2 --no-edit
git merge feat/your-feature-3 --no-edit
```

**Key principles:**

- **Prefix with `local/`** - signals this branch should never be pushed upstream
- **Base on `upstream/development`** - matches where PRs are merged
- **Use merge, not rebase** - easier to recreate when upstream changes
- **Recreate rather than update** - when upstream/development changes significantly, it's cleaner to recreate the integration branch from scratch:

```bash
# When upstream changes, recreate the integration branch
git checkout local/integration
git reset --hard upstream/development

# Re-merge your feature branches
git merge feat/your-feature-1 --no-edit
git merge feat/your-feature-2 --no-edit
```

This approach lets you test multiple features together locally without affecting the upstream repository or complicating your PR branches.

## Pull Request Guidelines

### Before Submitting

1. **Rebase against `upstream/development`** to ensure your changes are based on the latest code
2. **Run tests locally:** `go test ./...`
3. **Build the project:** `go build ./...`
4. **Test your changes manually** with real AWS credentials if applicable

### PR Requirements

- **Target the `development` branch** - not `main`
- **One feature per PR** - keep PRs focused and reviewable
- **Include tests** - new features should have accompanying unit tests
- **Update documentation** - update README.md if adding new flags or features
- **Follow existing patterns** - match the code style and architecture of existing code

### PR Title Format

Use conventional commit style:

- `feat: add new feature description`
- `fix: resolve bug description`
- `docs: update documentation`
- `test: add tests for feature`
- `refactor: improve code structure`
- `ci: update CI/CD configuration`

### During Review

- **Respond to feedback** promptly and professionally
- **Rebase when requested** - maintainers may ask you to rebase against the latest `development`
- **Don't force push** after approval without notifying reviewers

## Code Style

### Go Conventions

- Follow standard Go conventions and `gofmt`
- Use meaningful variable and function names
- Keep functions focused and reasonably sized
- Add comments for exported functions and complex logic

### Project-Specific Patterns

- **Service interfaces** are defined in `types.go` files
- **Service implementations** are in `service.go` files
- **AWS clients** use the AWS SDK v2 patterns
- **Concurrent operations** use `errgroup` for coordination

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party packages
    "github.com/aws/aws-sdk-go-v2/aws"

    // Internal packages
    "github.com/elC0mpa/aws-doctor/model"
)
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./service/ec2/...
```

### Writing Tests

- Use table-driven tests for comprehensive coverage
- Test edge cases and error conditions
- Don't modify production code solely for testability unless discussed
- Place test files alongside the code they test (`service.go` → `service_test.go`)

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "descriptive_test_case_name",
            input: ...,
            want:  ...,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            // assertions
        })
    }
}
```

## Commit Messages

### Format

```
type: short description

Longer description if needed, explaining the why
behind the change, not just what changed.

Co-Authored-By: Name <email> (if applicable)
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `test` - Adding or updating tests
- `refactor` - Code change that neither fixes a bug nor adds a feature
- `style` - Formatting, missing semicolons, etc.
- `ci` - CI/CD changes
- `chore` - Maintenance tasks

## Questions?

If you have questions or need help:

1. Check existing issues and PRs for similar topics
2. Open a new issue for discussion
3. Reach out to maintainers on PR comments

Thank you for contributing!
