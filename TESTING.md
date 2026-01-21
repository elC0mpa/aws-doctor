# Testing Strategy

This document outlines the testing approach for aws-doctor and discusses alternatives for testing AWS service interactions.

## Current Approach

The project currently uses pure unit tests for helper functions in the `utils/` package. These tests cover:
- Cost parsing and formatting
- Table generation
- JSON output formatting

## Testing AWS Service Interactions

Testing code that interacts with AWS services presents unique challenges. Here are the alternatives being considered:

### 1. gomock (Recommended)

Generate mocks from interfaces automatically. The AWS SDK interfaces are large but gomock handles this well.

```go
//go:generate mockgen -destination=mocks/mock_ec2.go -package=mocks github.com/aws/aws-sdk-go-v2/service/ec2 Client
```

**Pros:**
- Auto-generates mocks from interfaces
- Type-safe
- Well-maintained

**Cons:**
- Generated code can be verbose for large interfaces

### 2. testify/mock

More flexible mocking with assertion helpers.

```go
type MockEC2Client struct {
    mock.Mock
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
    args := m.Called(ctx, params)
    return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}
```

**Pros:**
- Flexible and easy to use
- Good assertion helpers
- Popular in the Go community

**Cons:**
- Manual mock implementation required
- Less type-safe than gomock

### 3. aws-sdk-go-v2-mock

AWS-specific mocking library designed for the SDK v2.

**Pros:**
- Designed specifically for AWS SDK v2
- Familiar patterns for AWS developers

**Cons:**
- Less community adoption
- May lag behind SDK updates

### 4. Integration Tests with LocalStack

Skip unit tests for AWS service calls and rely on integration tests with real AWS (or LocalStack).

```yaml
# docker-compose.yml for LocalStack
services:
  localstack:
    image: localstack/localstack
    ports:
      - "4566:4566"
    environment:
      - SERVICES=ec2,sts,ce,route53,elasticloadbalancingv2
```

**Pros:**
- Tests against real AWS API behavior
- Catches integration issues
- No mock maintenance

**Cons:**
- Slower test execution
- Requires Docker for local development
- LocalStack may not support all AWS features

## Recommended Approach

A hybrid approach is recommended:

1. **Unit tests with gomock** for business logic and service layer functions
2. **Integration tests with LocalStack** for end-to-end validation (optional, in CI)
3. **Pure unit tests** for utility functions (current approach)

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## Contributing Tests

When adding new features:

1. Add unit tests for any new utility functions
2. Consider adding mocked tests for service layer changes
3. Update this document if the testing strategy evolves
