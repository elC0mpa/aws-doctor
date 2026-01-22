# Testing Guide

This document explains how to write and run tests for aws-doctor.

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./service/orchestrator/...
```

## Test Structure

### Pure Unit Tests

Located in `utils/*_test.go` files. Test helper functions that don't require mocking:
- Cost parsing and formatting
- Table generation
- JSON output formatting

### Mocked Service Tests

Located in `service/orchestrator/service_test.go`. Use testify/mock to test service interactions.

#### Mock Implementations

Mocks are in `mocks/services.go` and implement the service interfaces:

- `MockSTSService` - mocks `awssts.STSService`
- `MockCostService` - mocks `awscostexplorer.CostService`
- `MockEC2Service` - mocks `awsec2.EC2Service`
- `MockELBService` - mocks `elb.ELBService`

#### Writing Mocked Tests

```go
func TestExample(t *testing.T) {
    // Create mocks
    mockSTS := new(mocks.MockSTSService)
    mockCost := new(mocks.MockCostService)
    mockEC2 := new(mocks.MockEC2Service)
    mockELB := new(mocks.MockELBService)

    // Setup expectations
    mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
        Account: aws.String("123456789012"),
    }, nil)

    // Create service with mocks
    svc := orchestrator.NewService(mockSTS, mockCost, mockEC2, mockELB)

    // Execute and assert
    err := svc.Orchestrate(model.Flags{Output: "json"})
    assert.NoError(t, err)
    mockSTS.AssertExpectations(t)
}
```

#### Testing Error Paths

```go
mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(nil, errors.New("API error"))
```

## Adding New Tests

When adding new features:

1. **Utility functions**: Add pure unit tests in the corresponding `*_test.go` file
2. **Service methods**: Add mock method to `mocks/services.go` if needed
3. **Orchestrator changes**: Add tests in `service/orchestrator/service_test.go`

## Test Style

Use table-driven tests for comprehensive coverage:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"descriptive_case_name", input, expected, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test
        })
    }
}
```
