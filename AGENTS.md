# AI Agent Instructions for aws-doctor

This file provides instructions for AI coding agents working on this project. For human contributors, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Project Overview

aws-doctor is a Go CLI tool that provides AWS cost analysis and waste detection. It acts as a free alternative to AWS Trusted Advisor.

### Key Features

- Cost comparison between current and previous month
- 6-month trend analysis
- Waste detection (unused EIPs, EBS volumes, stopped instances, load balancers, idle NAT Gateways, idle SageMaker real-time inference endpoints, etc.)
- Region-aware cost estimates: `service/pricing` fetches rates from the AWS Pricing API (always via the `us-east-1` endpoint, filtered by the caller's region) at startup and caches them in memory. The `Calculate*` helpers prefer the cached rate and fall back to the hardcoded us-east-1 defaults in `constants.go`. Categories currently fetched from the Pricing API: EBS volumes, EIP, NAT Gateway, ALB/NLB, CLB, CloudWatch Logs, RDS instances/storage/snapshots, and SageMaker hosting (real-time inference) instances. To add a new category, see the "Adding a Pricing Category" section in [CONTRIBUTING.md](CONTRIBUTING.md).
- Startup banner uses ANSI truecolor; title color switches to AmazonOrange when a blue background is detected (Windows console attributes or `COLORFGBG` on Unix-like terminals), otherwise SkypeBlue. Override with `AWS_DOCTOR_BANNER_COLOR` (color name or ANSI code).

## Quick Reference

```bash
# Build
go build ./...

# Test
go test ./...

# Run locally
go run . help
go run . cost
go run . waste
go run . trend
```

## Architecture

```
aws-doctor/
|-- app.go                 # Main application entry, delegates to cmd
|-- cmd/                   # Cobra CLI commands (root, waste, trend, etc.)
|-- model/                 # Data structures and types
|-- service/
|   |-- aws_config/       # AWS configuration loading
|   |-- costexplorer/     # AWS Cost Explorer service
|   |-- ec2/              # EC2 service (EIPs, EBS, instances)
|   |-- elb/              # ELB service (load balancers)
|   |-- orchestrator/     # Workflow coordination
|   |-- output/           # Output rendering (table/json) + spinner control
|   |-- sts/              # AWS STS service
|   |-- update/           # Self-update workflow
|-- utils/                # Utility functions, table rendering
|-- mocks/                # Test doubles
|   |-- services/         # Internal service mocks (for orchestrator tests)
|   |-- awsinterfaces/    # AWS SDK client mocks (for service tests)
|-- assets/               # Logos and images
|-- demo/                 # Demo GIFs
```

### Key Flows

- `app.go` delegates execution to `cmd.Execute()`.
- `cmd/` package defines Cobra commands (e.g., `cmd/waste.go`), initializes AWS services, and invokes `service/orchestrator`.
- `service/orchestrator` executes the workflow logic by calling service methods based on the configured model.Flags.
- `service/output` chooses between table and JSON rendering and owns spinner stop.
- `service/update` handles updates (invoked by `cmd/update.go`).

### Service Pattern

Each service follows this pattern to enable Dependency Injection for testing:

- `types.go` - Interface definitions (Service and AWS Client) and struct types
- `service.go` - Implementation

```go
// types.go
// 1. Define interface for AWS client methods used
type SomeClientAPI interface {
    SomeMethod(ctx context.Context, params *Input, optFns ...func(*Options)) (*Output, error)
}

type service struct {
    client SomeClientAPI // Use interface, not concrete struct
}

// 2. Define service interface
type ServiceInterface interface {
    Method(ctx context.Context) (Result, error)
}

// service.go
func NewService(awsconfig aws.Config) ServiceInterface {
    client := someclient.NewFromConfig(awsconfig)
    return &service{client: client}
}
```

### Service Naming & Boundaries

- **Method Naming**: Always use the `Get` prefix for service methods that retrieve information (e.g., `GetIdleNATGateways`, `GetRDSWaste`).
- **Service Boundaries**: Keep resources in their appropriate service. For example, NAT Gateway detection logic must reside in the `vpc` service, even if it uses the EC2 client under the hood.
- **Day Thresholds**: Never hardcode day thresholds (idle days, stale days, etc.) inside individual services. These must be passed as parameters from the orchestrator and defined as constants in the `orchestrator` service.
- **Parameter Naming**: Use `awsconfig` as the parameter name for `aws.Config` in all `NewService` constructors.

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

Use `errgroup` for concurrent AWS API calls. For internal result collection, **always prefer named internal structs** over anonymous ones to improve readability and maintainability.

```go
type result struct {
    data Result
    err  error
}

g, ctx := errgroup.WithContext(ctx)
results := make([]result, count)
...
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

- **Pure Unit Tests:** `utils/*_test.go` (no mocking required)
- **Service Unit Tests:** `service/*package*/service_test.go` (mocks AWS clients via `mocks/awsinterfaces`)
- **Orchestration Tests:** `service/orchestrator/service_test.go` (mocks internal services via `mocks/services`)

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

1. **Model Type**: Add the appropriate struct in `model/` package (e.g., `model/ec2.go` for `KeyPairWasteInfo`).
2. **Client Interface**: Define any new AWS client methods needed in the `*ClientAPI` interface (e.g., `service/ec2/types.go`).
3. **Client Mock**: Update the corresponding mock in `mocks/awsinterfaces/` to implement the new client method.
4. **Service Method**: Implement the logic in the service file (e.g., `service/ec2/service.go`). Use paginators for all AWS APIs that support them.
5. **Service Interface**: Add the new method to the `Service` interface in `types.go`.
6. **Service Mock**: Update the service mock in `mocks/services/` to include the new method. **This is critical to avoid `go vet` and build failures in orchestrator tests.**
7. **Orchestrator**: 
   - Add the concurrent call in `service/orchestrator/service.go` within `wasteWorkflow`.
   - Update the `RenderWaste` call to pass the new data.
8. **Output Service**: 
   - Update `RenderWaste` in `service/output/service.go`.
   - Update `Service` and `Renderer` interfaces in `service/output/types.go`.
   - Update `realRenderer` implementation in `service/output/types.go`.
   - Update `MockOutputService` in `mocks/services/output_service.go` and `MockRenderer` in `mocks/renderers/renderer.go`.
9. **Utility Handlers**:
   - Add a display function in `utils/waste_table/waste_table.go` and update `DrawWasteTable` signature.
   - Add a JSON output type in `model/output.go` and update `utils/json_output/json_output.go`.
10. **Test Compliance**: **Update all existing test calls** in `service/orchestrator`, `service/output`, and `utils` when function signatures change. Run `go test ./...` frequently.
11. **Documentation**: Update the feature checklist in `README.md`.
12. **Validation**: Run `go vet ./...` and `golangci-lint run` to ensure no regressions or interface mismatches were introduced.

### Adding a New Command or Flag

The CLI uses the Cobra framework.
1. **New Command**: Create a new file in `cmd/` (e.g., `cmd/myfeature.go`), define a `*cobra.Command`, and add it to `rootCmd` in its `init()` function.
2. **New Flag**: Add persistent flags to `rootCmd` in `cmd/root.go` for global flags, or local flags to specific commands.
3. Handle execution logic in the command's `RunE` method (often by instantiating the orchestrator and calling `orch.Orchestrate(flags)`).
4. Update `model.Flags` struct if you need to pass new flag states into the orchestrator.
5. Update `README.md` documentation.

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
