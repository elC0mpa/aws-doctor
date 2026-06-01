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
- **Update documentation** - update `README.md` (features and command lists) and the Docs Site (`docs/content/`) if adding new flags, waste checks, or features. Ensure new flags and arguments are added to the 'Selective Scanning' and 'Configuration Flags' tables in `docs/content/docs/waste-detection/_index.md` and `_index.es.md`. For developers and AI agents, also ensure `AGENTS.md` is updated if architectural patterns change.
- **Follow existing patterns** - match the code style and architecture of existing code

### Docs Site Card Grids

The "Instant Infrastructure Audit" section on the home page (`docs/content/_index.md` / `_index.es.md`) and the "Categories of Detection" grid on the waste-detection index (`docs/content/docs/waste-detection/_index.md` / `_index.es.md`) must mirror each other: every waste-detection category card must appear in both grids, and vice versa. When adding or removing a waste-detection category, update both pages.

Rules for these grids:
- **No duplicate links**: Two cards in the same grid must never point to the same page or anchor. If a sub-feature (e.g., Lambda) lives on an existing category page (e.g., Compute), mention it in that category's subtitle instead of creating a separate card.
- **One card per docs page**: Each waste-detection docs page (`compute.md`, `databases.md`, `storage.md`, `networking.md`, `machine-learning.md`, `configuration.md`, etc.) maps to exactly one card in each grid.

### IAM Permissions Style

On the docs site, IAM permissions must be expressed using inline callouts or plain text with backtick-formatted action names (e.g., `secretsmanager:ListSecrets`). Never use JSON policy blocks to show required permissions. Follow the pattern used in each waste-detection category page:

```
{{</* callout type="info" */>}}
**Permissions Required**: `action:One`, `action:Two`.
{{</* /callout */>}}
```

### PR Title Format

Use [Conventional Commits](https://www.conventionalcommits.org/) style:

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
- **Method Naming**: Service methods that retrieve data must use the `Get` prefix (e.g., `GetIdleNATGateways`, `GetRDSWaste`).
- **Day Thresholds**: All time-based thresholds for waste checks (e.g., idle days, stale days) must be parameterized in the service methods and defined as constants in the `service/orchestrator` package to centralize configuration.
- **Parameter Naming**: Use `awsconfig` consistently for `aws.Config` parameters in `NewService` constructors.
- **Struct Usage**: Prefer named internal structs over anonymous structs for complex return types or concurrent result collection to improve readability.
- **AWS clients** use the AWS SDK v2 patterns
- **Concurrent operations** use `errgroup` for coordination

### Adding a Pricing Category

Cost estimates for waste detection live in `service/pricing`. Each category fetches its rate from the AWS Pricing API at startup (filtered to the caller's region) and falls back to a hardcoded us-east-1 constant when the API call fails or returns no match. To add a new category:

1. **Add a category constant** to the `category*` block in `service/pricing/constants.go`, plus a default rate constant (e.g. `XxxCostPerMonth`) used as the fallback.
2. **Add a `fetch(...)` call** inside `LoadRegionRates` in `service/pricing/service.go`. Provide the AWS `serviceCode`, the `productFamily` term match, any other filters needed to narrow to the specific SKU, and an `extract` function that returns the cache-key variant (or `""` for single-rate categories) from the SKU attributes. Verify the attribute names against an actual `aws pricing get-products --service-code <SERVICE>` response — the Pricing API uses different keys (`instanceType`, `instanceName`, `volumeApiName`, `usagetype`) across services.
3. **Wire it into a `Calculate*` method** on `service` in `service/pricing/service.go`. Call `s.lookupMonthly(priceKey(category, variant), hoursPerMonth)` first when the Pricing API returns an hourly rate (e.g. compute instances, NAT, EIP), or `s.lookupMonthly(..., 0)` when it returns an already-monthly per-GB rate (e.g. EBS, CloudWatch Logs). Fall back to the hardcoded constant or table when the lookup misses. The cached value in `s.prices` is always stored as the raw `pricePerUnit` from the API, and `lookupMonthly`'s second argument is the multiplier applied on read.
4. **Expose the method** on the `Service` interface in `service/pricing/types.go` if it is a new public method.
5. **Add unit tests** in `service/pricing/service_test.go` covering both the cached path (use `s.setPrice(...)` to seed the cache) and the fallback path. Add a `TestLoadRegionRates_*` case using `mocks.MockPricingClient` with a representative price-list JSON document so the attribute names stay locked in.

Keep the fallback table around — it is used both when the Pricing API is unreachable and when a specific SKU is missing from the response.

### Adding a New Waste Detection Type

1. **Model Type**: Add the appropriate struct in `model/` package (e.g., `model/ec2.go` for `KeyPairWasteInfo`).
2. **Client Interface**: Define any new AWS client methods needed in the `*ClientAPI` interface (e.g., `service/ec2/types.go`).
3. **Client Mock**: Update the corresponding mock in `mocks/awsinterfaces/` to implement the new client method.
4. **Service Method**: Implement the logic in the service file (e.g., `service/ec2/service.go`). Use paginators for all AWS APIs that support them.
5. **Service Interface**: Add the new method to the `Service` interface in `types.go`.
6. **Service Mock**: Update the service mock in `mocks/services/` to include the new method.
7. **Analyzer Registration**: 
   - Ensure the service implements the `analyzer.WasteAnalyzer` interface (`Analyze` and `Name`).
   - If adding a completely new service, register it in the orchestrator builder (e.g., `cmd/root.go`, `cmd/waste.go`) via the `registry.Register()` method, and add its tab name mapping in `service/orchestrator/service_waste.go`.
   - If adding a new check inside an existing service, simply add the new method call to the concurrent `errgroup` inside the service's `Analyze` method. No orchestrator changes needed!
8. **Output Service**: 
   - Update `model.RenderWasteInput` in `model/waste.go` to include the new slice, and update its `Merge()` method.
   - Update `RenderWaste` signatures if necessary.
9. **Utility Handlers**:
   - Add a display function in `utils/waste_table/waste_table.go` and update the `RenderScopeTable` switch statement for the new scope (if applicable).
   - Add a JSON output type in `model/output.go` and update `utils/json_output/json_output.go`.
10. **Test Compliance**: **Update all existing test calls** in the service tests and `utils` when function signatures change. Run `go test ./...` frequently.
11. **Documentation**:
    - Update the feature checklist and 'Selective scanning' command list in `README.md`.
    - Add the new check to the 'Selective Scanning' table in `docs/content/docs/waste-detection/_index.md` and `_index.es.md`.
    - Create the dedicated documentation page for the category (e.g., `security.md` and `security.es.md`), **but only if the category doesn't exist already**.
    - Add the feature card to the Docs Site grids following the "Docs Site Card Grids" rules.
12. **Validation**: Run `go vet ./...` and `golangci-lint run` to ensure no regressions or interface mismatches were introduced.

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
- Use mocks for AWS services (see `mocks/` directory)
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

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

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
