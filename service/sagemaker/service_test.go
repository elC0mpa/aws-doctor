package sagemaker

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elC0mpa/aws-doctor/model"
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
		EndpointStatus:     smtypes.EndpointStatusInService,
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
	var (
		ctx          = context.Background()
		client       = new(awsinterfaces.MockSageMakerClient)
		cw           = new(mockCWMetricsService)
		pricingSvc   = new(services.MockPricingService)
		name         = "idle-endpoint"
		cfgName      = "idle-cfg"
		instanceType = "ml.t2.medium"
	)

	client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String(name), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	wireEndpoint(client, name, cfgName, instanceType)
	cw.On("SageMakerVariantInvocations", mock.Anything, name, "AllTraffic", 14).Return(0.0, nil)
	pricingSvc.On("CalculateSageMakerEndpointMonthlyCost", mock.Anything).Return(46.72)

	svc := &service{client: client, cwService: cw, pricingService: pricingSvc}
	results, err := svc.GetIdleEndpoints(ctx, 14)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, name, results[0].EndpointName)
	assert.Equal(t, 46.72, results[0].EstimatedMonthlyCost)
}

func TestGetIdleEndpoints_ActiveEndpointSkipped(t *testing.T) {
	var (
		ctx          = context.Background()
		client       = new(awsinterfaces.MockSageMakerClient)
		cw           = new(mockCWMetricsService)
		pricingSvc   = new(services.MockPricingService)
		name         = "active-endpoint"
		cfgName      = "active-cfg"
		instanceType = "ml.t2.medium"
	)

	client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
		Endpoints: []smtypes.EndpointSummary{
			{EndpointName: aws.String(name), EndpointStatus: smtypes.EndpointStatusInService},
		},
	}, nil)

	wireEndpoint(client, name, cfgName, instanceType)
	cw.On("SageMakerVariantInvocations", mock.Anything, name, "AllTraffic", 14).Return(100.0, nil)

	svc := &service{client: client, cwService: cw, pricingService: pricingSvc}
	results, err := svc.GetIdleEndpoints(ctx, 14)

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetIdleEndpoints_ListError(t *testing.T) {
	var (
		ctx    = context.Background()
		client = new(awsinterfaces.MockSageMakerClient)
	)

	client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return((*sm.ListEndpointsOutput)(nil), errors.New("list fail"))

	svc := &service{client: client}
	_, err := svc.GetIdleEndpoints(ctx, 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list fail")
}

func TestGetIdleEndpoints_CheckEndpointErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("DescribeEndpoint_Error", func(t *testing.T) {
		client := new(awsinterfaces.MockSageMakerClient)
		client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
			Endpoints: []smtypes.EndpointSummary{{EndpointName: aws.String("ep")}},
		}, nil)
		client.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("describe fail"))

		svc := &service{client: client}
		_, err := svc.GetIdleEndpoints(ctx, 14)
		assert.Error(t, err)
	})

	t.Run("DescribeEndpointConfig_Error", func(t *testing.T) {
		client := new(awsinterfaces.MockSageMakerClient)
		client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
			Endpoints: []smtypes.EndpointSummary{{EndpointName: aws.String("ep")}},
		}, nil)
		client.On("DescribeEndpoint", mock.Anything, mock.Anything, mock.Anything).Return(&sm.DescribeEndpointOutput{
			EndpointConfigName: aws.String("cfg"),
		}, nil)
		client.On("DescribeEndpointConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("cfg fail"))

		svc := &service{client: client}
		_, err := svc.GetIdleEndpoints(ctx, 14)
		assert.Error(t, err)
	})

	t.Run("CloudWatch_Error_ReturnsFalse", func(t *testing.T) {
		client := new(awsinterfaces.MockSageMakerClient)
		cw := new(mockCWMetricsService)

		wireEndpoint(client, "ep", "cfg", "ml.t2.medium")

		client.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(&sm.ListEndpointsOutput{
			Endpoints: []smtypes.EndpointSummary{{EndpointName: aws.String("ep")}},
		}, nil)

		cw.On("SageMakerVariantInvocations", mock.Anything, "ep", "AllTraffic", 14).Return(0.0, errors.New("cw fail"))

		svc := &service{client: client, cwService: cw}
		results, err := svc.GetIdleEndpoints(ctx, 14)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg, nil, nil)
	assert.NotNil(t, svc)
}

func TestAnalyzerMethods(t *testing.T) {
	svc := &service{}

	if svc.Name() == "" {
		t.Error("Name() should not be empty")
	}

	if svc.TabName() == "" {
		t.Error("TabName() should not be empty")
	}
}

func TestService_Analyze(t *testing.T) {
	mockClient := new(awsinterfaces.MockSageMakerClient)
	svc := &service{client: mockClient}

	assert.Equal(t, "sagemaker", svc.Name())
	assert.Equal(t, "SageMaker", svc.TabName())

	mockClient.On("ListEndpoints", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("err"))

	res, err := svc.Analyze(context.Background(), model.Flags{SageMakerIdleDays: 90})
	assert.NoError(t, err)
	assert.Equal(t, "sagemaker", res.Scope)
}
