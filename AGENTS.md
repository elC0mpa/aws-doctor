# AI Agent Instructions for aws-doctor

This file provides instructions for AI coding agents working on this project. For human contributors, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Project Overview

aws-doctor is a Go CLI tool that provides AWS cost analysis and waste detection. It acts as a free alternative to AWS Trusted Advisor.

### Key Features
- Cost comparison between current and previous month
- 6-month trend analysis
- Waste detection (unused EIPs, EBS volumes, stopped instances, load balancers, etc.)

## Quick Reference

```bash
# Build
go build ./...

# Test
go test ./...

# Run locally
go run . --help
go run . --waste
go run . --trend
```

## Architecture

```
aws-doctor/
|-- app.go                 # Main application entry, flag parsing
|-- model/                 # Data structures and types
|-- service/
|   |-- aws_config/       # AWS configuration loading
|   |-- costexplorer/     # AWS Cost Explorer service
|   |-- ec2/              # EC2 service (EIPs, EBS, instances)
|   |-- elb/              # ELB service (load balancers)
|   |-- flag/             # CLI flag parsing
|   |-- orchestrator/     # Workflow coordination
|   |-- output/           # Output rendering (table/json) + spinner control
|   |-- sts/              # AWS STS service
|   |-- update/           # Self-update workflow
|-- utils/                # Utility functions, table rendering
|-- mocks/                # Test doubles for orchestrator tests
|-- assets/               # Logos and images
|-- demo/                 # Demo GIFs
```

### Key Flows

- `app.go` builds services, then delegates to `service/orchestrator`.
- `service/orchestrator` selects a workflow based on flags and calls service methods.
- `service/output` chooses between table and JSON rendering and owns spinner stop.
- `service/update` handles `--update`.

### Service Pattern

Each service follows this pattern:
- `types.go` - Interface definitions and struct types
- `service.go` - Implementation

```go
// types.go
type service struct {
    client *aws.Client
}

type ServiceInterface interface {
    Method(ctx context.Context) (Result, error)
}

// service.go
func NewService(cfg aws.Config) *service {
    return &service{client: aws.NewFromConfig(cfg)}
}

func (s *service) Method(ctx context.Context) (Result, error) {
    // implementation
}
```

## Git Workflow

### Critical Rules

1. **Always target `development` branch** for PRs, never `main`
2. **Always rebase against upstream** before pushing
3. **Fetch upstream frequently** to stay current

### Remote Setup

Contributors typically have:
- `origin` - their fork
- `upstream` - the original repo (elC0mpa/aws-doctor)

Note: Some may use different names. Adjust commands accordingly.

```bash
# Sync with upstream before work
git fetch upstream
git checkout development
git reset --hard upstream/development

# Create feature branch
git checkout -b feat/feature-name upstream/development

# Before PR or when requested by maintainer
git fetch upstream
git rebase upstream/development
git push origin feat/feature-name --force
```

## Code Guidelines

### Imports

Use import aliases for AWS SDK packages to avoid conflicts:
```go
import (
    elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
    elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)
```

### Concurrency

Use `errgroup` for concurrent AWS API calls:
```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    result, err = s.service.Method(ctx)
    return err
})

if err := g.Wait(); err != nil {
    return err
}
```

### Pagination

Use AWS SDK v2 paginators for APIs that return paginated results:
```go
paginator := elb.NewDescribeLoadBalancersPaginator(s.client, &elb.DescribeLoadBalancersInput{})
for paginator.HasMorePages() {
    output, err := paginator.NextPage(ctx)
    if err != nil {
        return nil, err
    }
    results = append(results, output.Items...)
}
```

### Error Handling

- Return errors to callers, don't log and continue
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Check for nil pointers before dereferencing AWS response fields

### Linting Compliance

The CI runs golangci-lint. Common issues to avoid:
- **S1017**: Use `strings.TrimPrefix(s, prefix)` directly instead of `if strings.HasPrefix(s, prefix) { s = strings.TrimPrefix(s, prefix) }`
- Remove unused imports (the build will fail)

## Testing

### Current Approach

Tests focus on pure functions that don't require AWS mocking:
- `utils/ec2_test.go` - Date parsing
- `service/ec2/service_test.go` - Resource type detection
- `service/costexplorer/service_test.go` - Date helpers, filtering

### Test Style

Use table-driven tests:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"case_name", input, expected, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test
        })
    }
}
```

## Common Tasks

## Documentation Maintenance (Required)

Any change that affects behavior, flags, outputs, workflows, supported AWS resources, build/test steps, or architecture must be reflected in the documentation. Agents must update the relevant files as part of the same change:

- `AGENTS.md` for agent guidance, architecture, workflows, and contribution rules.
- `README.md` for user-facing behavior, flags, features, and roadmap/checklists.
- `CONTRIBUTING.md` and `TESTING.md` for contributor workflow and test guidance.

If a change makes documentation inaccurate or incomplete, treat the documentation update as mandatory and do it in the same patch/PR.

### Adding a New Waste Detection Type

1. Add model type in `model/` package (e.g., `model/ec2.go` for `SnapshotWasteInfo`)
2. Add method to `service/ec2/service.go` (or appropriate service)
3. Update interface in `service/ec2/types.go`
4. Add concurrent call in `service/orchestrator/service.go` wasteWorkflow
5. Update `service/output/service.go` to pass the new data into `RenderWaste`
6. Add display function in `utils/waste_table.go` (update `DrawWasteTable` signature)
7. Add JSON output type in `model/output.go` and update `utils/json_output.go`
8. **Update all test calls** when function signatures change (e.g., `DrawWasteTable`, `OutputWasteJSON`)
9. Update README.md checklist
10. Add tests for any pure helper functions

### Adding a New CLI Flag

1. Add flag definition in `service/flag/service.go`
2. Add field to `model.Flags` struct
3. Handle flag in `service/orchestrator/service.go`
4. Update README.md documentation

## PR Checklist

Before submitting:
- [ ] Rebased against upstream `development` branch
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] New features have tests (for testable code)
- [ ] README.md updated if adding flags/features
- [ ] PR targets `development` branch (not `main`)

After pushing:
- [ ] CI passes (build, lint, tests on Go 1.23 and 1.24)
- [ ] Address any golangci-lint warnings

## PR Best Practices

- **Keep PRs focused** - one feature/fix per PR
- **Maintainers may ask to split PRs** - if a PR has parts with different dependencies, be prepared to split it
- **Rebase when asked** - maintainers may request rebasing after upstream changes
- **CI must pass** - fix any build, lint, or test failures before requesting review

## Don't

- Don't modify production code solely to make it testable (discuss first)
- Don't add interfaces for mocking without maintainer approval
- Don't commit AWS credentials or sensitive data
- Don't target `main` branch for PRs
- Don't force push to shared branches after approval
