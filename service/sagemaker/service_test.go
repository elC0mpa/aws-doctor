package sagemaker

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCWMetricsService struct {
	mock.Mock
}

func (m *mockCWMetricsService) SageMakerVariantInvocations(ctx context.Context, endpointName, variantName string, days int) (float64, error) {
	args := m.Called(ctx, endpointName, variantName, days)

	return args.Get(0).(float64), args.Error(1)
}

// wireEndpoint builds a minimal endpoint setup: a DescribeEndpoint response and a matching
// DescribeEndpointConfig response with one AllTraffic production variant.
func wireEndpoint(client *awsinterfaces.MockSageMakerClient, name, cfgName, instanceType string) {
	const (
		variantName = "AllTraffic"
		count       = int32(1)
	)
	client.On("DescribeEndpoint", mock.Anything, mock.MatchedBy(func(in *sm.DescribeEndpointInput) bool {
		return aws.ToString(in.EndpointName) == name
	}), mock.Anything).Return(&sm.DescribeEndpointOutput{
		EndpointName:       aws.String(name),
		EndpointConfigName: aws.String(cfgName),
		ProductionVariants: []smtypes.ProductionVariantSummary{
			{
				VariantName:          aws.String(variantName),
				CurrentInstanceCount: aws.Int32(count),
			},
		},
	}, nil)

	client.On("DescribeEndpointConfig", mock.Anything, mock.MatchedBy(func(in *sm.DescribeEndpointConfigInput) bool {
		return aws.ToString(in.EndpointConfigName) == cfgName
	}), mock.Anything).Return(&sm.DescribeEndpointConfigOutput{
		ProductionVariants: []smtypes.ProductionVariant{
			{
				VariantName:  aws.String(variantName),
				InstanceType: smtypes.ProductionVariantInstanceType(instanceType),
			},
		},
	}, nil)
}

func TestGetIdleEndpoints_IdleEndpointReported(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{
				EndpointName:   aws.String("idle-ep"),
				EndpointArn:    aws.String("arn:aws:sagemaker:us-east-1:111111111111:endpoint/idle-ep"),
				EndpointStatus: smtypes.EndpointStatusInService,
			},
		},
	}, nil)

	wireEndpoint(mockClient, "idle-ep", "idle-ep-cfg", "ml.m5.xlarge")

	mockCW.On("SageMakerVariantInvocations", mock.Anything, "idle-ep", "AllTraffic", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "idle-ep", result[0].EndpointName)
	assert.Equal(t, 14, result[0].DaysChecked)
	assert.InDelta(t, 203.82, result[0].EstimatedMonthlyCost, 1e-6)
	assert.Len(t, result[0].Variants, 1)
	assert.Equal(t, "ml.m5.xlarge", result[0].Variants[0].InstanceType)
}

func TestGetIdleEndpoints_ActiveEndpointExcluded(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{
				EndpointName:   aws.String("active-ep"),
				EndpointStatus: smtypes.EndpointStatusInService,
			},
		},
	}, nil)

	wireEndpoint(mockClient, "active-ep", "active-ep-cfg", "ml.m5.large")

	mockCW.On("SageMakerVariantInvocations", mock.Anything, "active-ep", "AllTraffic", 14).Return(float64(42), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetIdleEndpoints_EndpointIdleOnlyWhenAllVariantsIdle(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{
				EndpointName:   aws.String("multi-variant-ep"),
				EndpointStatus: smtypes.EndpointStatusInService,
			},
		},
	}, nil)

	mockClient.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointOutput{
		EndpointName:       aws.String("multi-variant-ep"),
		EndpointConfigName: aws.String("cfg"),
		ProductionVariants: []smtypes.ProductionVariantSummary{
			{VariantName: aws.String("v1"), CurrentInstanceCount: aws.Int32(1)},
			{VariantName: aws.String("v2"), CurrentInstanceCount: aws.Int32(1)},
		},
	}, nil)

	mockClient.On("DescribeEndpointConfig", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointConfigOutput{
		ProductionVariants: []smtypes.ProductionVariant{
			{VariantName: aws.String("v1"), InstanceType: smtypes.ProductionVariantInstanceTypeMlM5Large},
			{VariantName: aws.String("v2"), InstanceType: smtypes.ProductionVariantInstanceTypeMlM5Large},
		},
	}, nil)

	// v1 has traffic, v2 is idle — endpoint overall should NOT be flagged.
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "multi-variant-ep", "v1", 14).Return(float64(10), nil)
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "multi-variant-ep", "v2", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Empty(t, result, "endpoint with ANY variant traffic must not be flagged idle")
}

func TestGetIdleEndpoints_NoEndpoints(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{},
	}, nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetIdleEndpoints_ListEndpointsError(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))

	_, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list SageMaker endpoints")
}

func TestGetIdleEndpoints_ZeroIdleDaysUsesDefault(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{
				EndpointName:   aws.String("idle-ep"),
				EndpointStatus: smtypes.EndpointStatusInService,
			},
		},
	}, nil)

	wireEndpoint(mockClient, "idle-ep", "idle-ep-cfg", "ml.t3.medium")

	// Days argument is 0; default 14 must be used.
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "idle-ep", "AllTraffic", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 14, result[0].DaysChecked)
}

func TestGetIdleEndpoints_CloudWatchErrorDropsEndpointWithoutAborting(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("metric-err-ep"), EndpointStatus: smtypes.EndpointStatusInService},
			{EndpointName: aws.String("idle-ep"), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	wireEndpoint(mockClient, "metric-err-ep", "metric-err-cfg", "ml.m5.xlarge")
	wireEndpoint(mockClient, "idle-ep", "idle-cfg", "ml.m5.xlarge")

	// A CloudWatch error on one endpoint must not abort the scan and must not surface that
	// endpoint as idle (we don't know its state, so dropping is the safer behavior).
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "metric-err-ep", "AllTraffic", 14).Return(float64(0), errors.New("cw throttled"))
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "idle-ep", "AllTraffic", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err, "a CloudWatch metric error must not fail the whole scan")
	assert.Len(t, result, 1)
	assert.Equal(t, "idle-ep", result[0].EndpointName)
}

func TestGetIdleEndpoints_DescribeEndpointErrorPropagates(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("bad-ep"), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	mockClient.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return((*sm.DescribeEndpointOutput)(nil), errors.New("access denied"))

	_, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "describe endpoint bad-ep")
}

func TestGetIdleEndpoints_DescribeEndpointConfigErrorPropagates(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("ep"), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	mockClient.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointOutput{
		EndpointName:       aws.String("ep"),
		EndpointConfigName: aws.String("missing-cfg"),
		ProductionVariants: []smtypes.ProductionVariantSummary{
			{VariantName: aws.String("AllTraffic"), CurrentInstanceCount: aws.Int32(1)},
		},
	}, nil)

	mockClient.On("DescribeEndpointConfig", mock.Anything, mock.Anything, mock.Anything).Return((*sm.DescribeEndpointConfigOutput)(nil), errors.New("config not found"))

	_, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "describe endpoint config missing-cfg")
}

func TestGetIdleEndpoints_PaginatesListEndpoints(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	// First page returns ep-1 with a NextToken.
	mockClient.On("ListEndpoints", mock.Anything, mock.MatchedBy(func(in *sm.ListEndpointsInput) bool {
		return in.NextToken == nil && in.StatusEquals == smtypes.EndpointStatusInService
	}), mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("ep-1"), EndpointStatus: smtypes.EndpointStatusInService},
		},
		NextToken: aws.String("page-2"),
	}, nil).Once()

	// Second page returns ep-2 with no NextToken.
	mockClient.On("ListEndpoints", mock.Anything, mock.MatchedBy(func(in *sm.ListEndpointsInput) bool {
		return aws.ToString(in.NextToken) == "page-2"
	}), mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("ep-2"), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil).Once()

	wireEndpoint(mockClient, "ep-1", "cfg-1", "ml.m5.large")
	wireEndpoint(mockClient, "ep-2", "cfg-2", "ml.m5.large")

	mockCW.On("SageMakerVariantInvocations", mock.Anything, "ep-1", "AllTraffic", 14).Return(float64(0), nil)
	mockCW.On("SageMakerVariantInvocations", mock.Anything, "ep-2", "AllTraffic", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Len(t, result, 2, "both pages must be scanned")
	names := []string{result[0].EndpointName, result[1].EndpointName}
	assert.ElementsMatch(t, []string{"ep-1", "ep-2"}, names)
	mockClient.AssertExpectations(t)
}

func TestGetIdleEndpoints_EmptyVariantsExcluded(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String("novariants-ep"), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	mockClient.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointOutput{
		EndpointName:       aws.String("novariants-ep"),
		EndpointConfigName: aws.String("cfg"),
		ProductionVariants: []smtypes.ProductionVariantSummary{},
	}, nil)

	mockClient.On("DescribeEndpointConfig", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointConfigOutput{
		ProductionVariants: []smtypes.ProductionVariant{},
	}, nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Empty(t, result, "endpoint with no production variants should not be reported")
}

func TestGetIdleEndpoints_UnknownInstanceTypeStillReportsEndpoint(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	mockCW := new(mockCWMetricsService)

	s := &service{client: mockClient, cwService: mockCW}

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{
				EndpointName:   aws.String("unknown-type-ep"),
				EndpointStatus: smtypes.EndpointStatusInService,
			},
		},
	}, nil)

	// Use an instance type string not in the fallback pricing table.
	wireEndpoint(mockClient, "unknown-type-ep", "cfg", "ml.somefutureinstance.xlarge")

	mockCW.On("SageMakerVariantInvocations", mock.Anything, "unknown-type-ep", "AllTraffic", 14).Return(float64(0), nil)

	result, err := s.GetIdleEndpoints(context.Background(), 14)

	assert.NoError(t, err)
	assert.Len(t, result, 1, "unknown instance types should not silently drop the endpoint from the report")
	assert.Equal(t, float64(0), result[0].EstimatedMonthlyCost, "unknown types contribute 0 to cost but the endpoint is still surfaced")
}
