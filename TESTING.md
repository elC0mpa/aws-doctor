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

Located in `service/*package*/service_test.go`. We use `testify/mock` and Dependency Injection (DI) to test service logic in isolation.

#### Mock Directory Structure

The `mocks/` directory is organized into two subdirectories:

1.  **`mocks/services/`**: Mocks of our internal application services (e.g., `MockCostService`, `MockEC2Service`). These are used when testing higher-level components like the `orchestrator`.
2.  **`mocks/awsinterfaces/`**: Mocks of external AWS SDK clients (e.g., `MockCostExplorerClient`). These are used when testing low-level service adapters (e.g., `service/costexplorer`).

#### Testing Pattern (Dependency Injection)

To make AWS services testable, we define an interface for the AWS client methods we use, rather than depending on the concrete struct.

**1. Define the Interface (in `service/pkg/types.go`):**

```go
// CostExplorerClientAPI is the interface for the AWS Cost Explorer client methods used by the service.
type CostExplorerClientAPI interface {
    GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error)
}

type service struct {
    client CostExplorerClientAPI
}
```

**2. Implement the Mock (in `mocks/awsinterfaces/pkg.go`):**

```go
type MockCostExplorerClient struct {
    mock.Mock
}

func (m *MockCostExplorerClient) GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error) {
    args := m.Called(ctx, params, optFns)
    return args.Get(0).(*costexplorer.GetCostAndUsageOutput), args.Error(1)
}
```

**3. Write the Test (in `service/pkg/service_test.go`):**

```go
func TestGetMonthCostsByService(t *testing.T) {
    // Create mock
    mockClient := new(awsinterfaces.MockCostExplorerClient)
    
    // Inject mock into service
    s := &service{client: mockClient}

    // Setup expectations
    mockClient.On("GetCostAndUsage", mock.Anything, mock.Anything, mock.Anything).Return(&costexplorer.GetCostAndUsageOutput{...}, nil)

    // Execute
    result, err := s.GetMonthCostsByService(context.Background(), time.Now())

    // Assert
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

## Adding New Tests

When adding new features:

1.  **Utility functions**: Add pure unit tests in the corresponding `*_test.go` file.
2.  **New AWS calls**:
    *   Add the method to the client interface in `types.go`.
    *   Update the mock in `mocks/awsinterfaces/`.
    *   Add unit tests in `service_test.go` mocking the new call.
3.  **Service methods**:
    *   Add the method to the internal service interface.
    *   Update the mock in `mocks/services/`.
4.  **Orchestrator changes**: Add tests in `service/orchestrator/service_test.go` using the service mocks.

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
