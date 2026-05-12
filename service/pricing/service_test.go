package pricing

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	awspricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateEBSMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback value", func(t *testing.T) {
		// gp2 fallback is 0.10
		cost := s.CalculateEBSMonthlyCost(100, types.VolumeTypeGp2)
		assert.Equal(t, 10.0, cost)
	})

	t.Run("cached value", func(t *testing.T) {
		s.setPrice(priceKey(categoryEBS, "gp3"), 0.05)
		cost := s.CalculateEBSMonthlyCost(100, types.VolumeTypeGp3)
		assert.Equal(t, 5.0, cost)
	})

	t.Run("unsupported type uses gp2 fallback", func(t *testing.T) {
		cost := s.CalculateEBSMonthlyCost(100, types.VolumeType("unknown"))
		assert.Equal(t, 10.0, cost)
	})
}

func TestCalculateEBSSnapshotMonthlyCost(t *testing.T) {
	s := &service{}
	cost := s.CalculateEBSSnapshotMonthlyCost(100)
	assert.Equal(t, 5.0, cost)
}

func TestCalculateEIPMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback value", func(t *testing.T) {
		cost := s.CalculateEIPMonthlyCost()
		assert.Equal(t, EIPCostPerMonth, cost)
	})

	t.Run("cached value", func(t *testing.T) {
		// Rate is per hour, we expect lookupMonthly to multiply by hoursPerMonth (730)
		s.setPrice(priceKey(categoryEIP, ""), 0.01)
		cost := s.CalculateEIPMonthlyCost()
		assert.Equal(t, 7.3, cost)
	})
}

func TestCalculateLoadBalancerMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("ALB fallback", func(t *testing.T) {
		cost := s.CalculateLoadBalancerMonthlyCost(elbtypes.LoadBalancerTypeEnumApplication)
		assert.Equal(t, ALBCostPerMonth, cost)
	})

	t.Run("CLB fallback", func(t *testing.T) {
		cost := s.CalculateLoadBalancerMonthlyCost(elbtypes.LoadBalancerTypeEnum("classic"))
		assert.Equal(t, CLBCostPerMonth, cost)
	})

	t.Run("CLB cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryLBClassic, ""), 0.04)
		cost := s.CalculateLoadBalancerMonthlyCost(elbtypes.LoadBalancerTypeEnum("classic"))
		assert.Equal(t, 0.04*hoursPerMonth, cost)
	})

	t.Run("ALB cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryLBApp, "application"), 0.03)
		cost := s.CalculateLoadBalancerMonthlyCost(elbtypes.LoadBalancerTypeEnumApplication)
		assert.Equal(t, 0.03*hoursPerMonth, cost)
	})

	t.Run("unknown type fallback", func(t *testing.T) {
		cost := s.CalculateLoadBalancerMonthlyCost(elbtypes.LoadBalancerTypeEnum("unknown"))
		assert.Equal(t, ALBCostPerMonth, cost)
	})
}

func TestCalculateCloudWatchLogsMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback", func(t *testing.T) {
		// 100 GB * 0.03
		cost := s.CalculateCloudWatchLogsMonthlyCost(100 * 1024 * 1024 * 1024)
		assert.InDelta(t, 3.0, cost, 0.001)
	})

	t.Run("cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryCWLogs, ""), 0.05)
		cost := s.CalculateCloudWatchLogsMonthlyCost(100 * 1024 * 1024 * 1024)
		assert.InDelta(t, 5.0, cost, 0.001)
	})
}

func TestCalculateRDSInstanceMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("Single-AZ fallback", func(t *testing.T) {
		cost := s.CalculateRDSInstanceMonthlyCost(100, false)
		assert.Equal(t, 100*RDSStorageCostPerGBMonth, cost)
	})

	t.Run("Multi-AZ cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryRDSStorage, ""), 0.20)
		cost := s.CalculateRDSInstanceMonthlyCost(100, true)
		assert.Equal(t, 40.0, cost)
	})
}

func TestCalculateRDSIdleInstanceMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback", func(t *testing.T) {
		// db.t3.micro is 12.41, storage is 0.115/GB
		cost := s.CalculateRDSIdleInstanceMonthlyCost("db.t3.micro", 10, false)
		assert.InDelta(t, 12.41+(10*0.115), cost, 0.001)
	})

	t.Run("cached Multi-AZ", func(t *testing.T) {
		s.setPrice(priceKey(categoryRDSInstance, "db.t3.medium"), 0.068) // hourly
		s.setPrice(priceKey(categoryRDSStorage, ""), 0.10)               // monthly
		// (0.068 * 730) + (100 * 0.10) = 49.64 + 10 = 59.64. Multi-AZ = 119.28
		cost := s.CalculateRDSIdleInstanceMonthlyCost("db.t3.medium", 100, true)
		assert.InDelta(t, 119.28, cost, 0.001)
	})
}

func TestCalculateSageMakerEndpointMonthlyCost(t *testing.T) {
	t.Run("fallback table", func(t *testing.T) {
		s := &service{prices: make(map[string]float64)}
		variants := []model.SageMakerVariant{
			{InstanceType: "ml.t2.medium", InstanceCount: 2},
		}
		// 46.72 * 2 = 93.44
		cost := s.CalculateSageMakerEndpointMonthlyCost(variants)
		assert.Equal(t, 93.44, cost)
	})

	t.Run("cached value preferred over fallback", func(t *testing.T) {
		s := &service{prices: make(map[string]float64)}
		// hourly rate; lookupMonthly multiplies by hoursPerMonth (730)
		s.setPrice(priceKey(categorySageMakerHosting, "ml.m5.xlarge"), 0.30)

		variants := []model.SageMakerVariant{
			{InstanceType: "ml.m5.xlarge", InstanceCount: 1},
		}
		assert.InDelta(t, 0.30*hoursPerMonth, s.CalculateSageMakerEndpointMonthlyCost(variants), 0.001)
	})

	t.Run("mixed cached and fallback variants", func(t *testing.T) {
		s := &service{prices: make(map[string]float64)}
		s.setPrice(priceKey(categorySageMakerHosting, "ml.m5.xlarge"), 0.30)

		variants := []model.SageMakerVariant{
			{InstanceType: "ml.m5.xlarge", InstanceCount: 1}, // cached: 0.30 * 730 = 219.0
			{InstanceType: "ml.t2.medium", InstanceCount: 1}, // fallback: 46.72
		}
		assert.InDelta(t, 0.30*hoursPerMonth+46.72, s.CalculateSageMakerEndpointMonthlyCost(variants), 0.001)
	})

	t.Run("unknown instance type contributes zero", func(t *testing.T) {
		s := &service{prices: make(map[string]float64)}
		variants := []model.SageMakerVariant{
			{InstanceType: "ml.does-not-exist.xlarge", InstanceCount: 4},
		}
		assert.Equal(t, 0.0, s.CalculateSageMakerEndpointMonthlyCost(variants))
	})
}

func TestLoadRegionRates_SageMakerHosting(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	sagemakerJSON := `{
		"product": { "attributes": { "instanceName": "ml.m5.xlarge", "component": "Hosting", "productFamily": "ML Instance" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.30" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{sagemakerJSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)

	price, ok := s.lookupMonthly(priceKey(categorySageMakerHosting, "ml.m5.xlarge"), 0)
	assert.True(t, ok)
	assert.Equal(t, 0.30, price)
}

func TestLoadRegionRates(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	validJSON := `{
		"product": { "attributes": { "volumeApiName": "gp3", "usagetype": "EBS:VolumeUsage.gp3" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.08" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{validJSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)

	price, ok := s.lookupMonthly(priceKey(categoryEBS, "gp3"), 0)
	assert.True(t, ok)
	assert.Equal(t, 0.08, price)
}

func TestLoadRegionRates_ECR(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	ecrJSON := `{
		"product": { "attributes": { "productFamily": "EC2 Container Registry", "usagetype": "USE1-TimedStorage-ByteHrs" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.10" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{ecrJSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)

	price, ok := s.lookupMonthly(priceKey(categoryECR, ""), 0)
	assert.True(t, ok)
	assert.Equal(t, 0.10, price)
}

func TestLoadRegionRates_Error(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return((*awspricing.GetProductsOutput)(nil), errors.New("api error"))

	err := s.LoadRegionRates(ctx)
	assert.Error(t, err)
}

func TestLoadRegionRates_ExtractionSkip(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	// JSON with missing attributes to trigger !ok in extract
	invalidAttrJSON := `{
		"product": { "attributes": { "wrongKey": "value" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.08" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{invalidAttrJSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)
	assert.Empty(t, s.prices)
}

func TestLoadRegionRates_LBUsageSkip(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	// JSON for LB but with usagetype that fails matchLBUsage (contains TS-)
	invalidLBJSON := `{
		"product": { "attributes": { "productFamily": "Load Balancer-Application", "usagetype": "USE1-TS-LoadBalancerUsage" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.02" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{invalidLBJSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)
	assert.Empty(t, s.prices)
}

func TestParsePriceListDocument(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantPrice float64
		wantOk    bool
	}{
		{
			name: "valid json",
			raw: `{
				"product": { "attributes": { "instanceType": "db.t3.medium" } },
				"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.068" } } } } } }
			}`,
			wantPrice: 0.068,
			wantOk:    true,
		},
		{
			name:      "invalid json",
			raw:       `{invalid}`,
			wantPrice: 0,
			wantOk:    false,
		},
		{
			name: "missing usd",
			raw: `{
				"product": { "attributes": {} },
				"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "EUR": "0.06" } } } } } }
			}`,
			wantPrice: 0,
			wantOk:    false,
		},
		{
			name: "invalid usd string",
			raw: `{
				"product": { "attributes": {} },
				"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "invalid" } } } } } }
			}`,
			wantPrice: 0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := parsePriceListDocument(tt.raw)
			assert.Equal(t, tt.wantOk, ok)

			if ok {
				assert.Equal(t, tt.wantPrice, entry.pricePerUnit)
			}
		})
	}
}

func TestCalculateEC2InstanceMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback table", func(t *testing.T) {
		// t3.medium fallback is 30.37
		assert.Equal(t, 30.37, s.CalculateEC2InstanceMonthlyCost("t3.medium"))
	})

	t.Run("cached hourly converted to monthly", func(t *testing.T) {
		s.setPrice(priceKey(categoryEC2Instance, "m5.large"), 0.10)
		assert.Equal(t, 0.10*hoursPerMonth, s.CalculateEC2InstanceMonthlyCost("m5.large"))
	})

	t.Run("unknown instance type returns zero", func(t *testing.T) {
		assert.Equal(t, 0.0, s.CalculateEC2InstanceMonthlyCost("foo.bar"))
	})
}

func TestLoadRegionRates_EC2Instance(t *testing.T) {
	ctx := context.Background()
	mockClient := new(awsinterfaces.MockPricingClient)
	s := &service{
		client: mockClient,
		prices: make(map[string]float64),
		region: "us-east-1",
	}

	ec2JSON := `{
		"product": { "attributes": { "instanceType": "m6i.large", "operatingSystem": "Linux", "tenancy": "Shared", "preInstalledSw": "NA", "capacitystatus": "Used" } },
		"terms": { "OnDemand": { "SKU": { "priceDimensions": { "DIM": { "pricePerUnit": { "USD": "0.096" } } } } } }
	}`

	mockClient.On("GetProducts", mock.Anything, mock.Anything, mock.Anything).Return(&awspricing.GetProductsOutput{
		PriceList: []string{ec2JSON},
	}, nil)

	err := s.LoadRegionRates(ctx)
	assert.NoError(t, err)

	price, ok := s.lookupMonthly(priceKey(categoryEC2Instance, "m6i.large"), 0)
	assert.True(t, ok)
	assert.Equal(t, 0.096, price)
}

func TestCalculateNATGatewayMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback", func(t *testing.T) {
		assert.Equal(t, NatGatewayCostPerMonth, s.CalculateNATGatewayMonthlyCost())
	})

	t.Run("cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryNAT, ""), 0.05)
		assert.Equal(t, 0.05*hoursPerMonth, s.CalculateNATGatewayMonthlyCost())
	})
}

func TestCalculateRDSSnapshotMonthlyCost(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	t.Run("fallback", func(t *testing.T) {
		assert.Equal(t, float64(100)*RDSSnapshotCostPerGBMonth, s.CalculateRDSSnapshotMonthlyCost(100))
	})

	t.Run("cached", func(t *testing.T) {
		s.setPrice(priceKey(categoryRDSSnapshot, ""), 0.15)
		assert.Equal(t, 15.0, s.CalculateRDSSnapshotMonthlyCost(100))
	})
}

func TestRDSInstanceComputeCost_Trimming(t *testing.T) {
	s := &service{
		prices: make(map[string]float64),
	}

	// db.m5.large -> m5.large
	s.setPrice(priceKey(categoryRDSInstance, "m5.large"), 0.10)
	cost := s.rdsInstanceComputeCost("db.m5.large")
	assert.Equal(t, 0.10*hoursPerMonth, cost)
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{Region: "us-west-2"}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
	s := svc.(*service)
	assert.Equal(t, "us-west-2", s.region)
}

func TestPriceKey(t *testing.T) {
	assert.Equal(t, "ebs:gp3", priceKey("ebs", "gp3"))
	assert.Equal(t, "nat", priceKey("nat", ""))
}
