package pricing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSageMakerEndpointMonthlyCost(t *testing.T) {
	tests := []struct {
		name     string
		variants []SageMakerVariantCost
		want     float64
	}{
		{
			name:     "empty variants returns zero",
			variants: nil,
			want:     0,
		},
		{
			name: "single known instance",
			variants: []SageMakerVariantCost{
				{InstanceType: "ml.m5.xlarge", InstanceCount: 1},
			},
			want: 203.82,
		},
		{
			name: "multiple instances scale linearly",
			variants: []SageMakerVariantCost{
				{InstanceType: "ml.m5.xlarge", InstanceCount: 3},
			},
			want: 611.46,
		},
		{
			name: "variants sum across types",
			variants: []SageMakerVariantCost{
				{InstanceType: "ml.m5.large", InstanceCount: 2},
				{InstanceType: "ml.g4dn.xlarge", InstanceCount: 1},
			},
			want: 203.82 + 538.72,
		},
		{
			name: "unknown instance contributes zero, known contributes normally",
			variants: []SageMakerVariantCost{
				{InstanceType: "ml.neverseen.xlarge", InstanceCount: 4},
				{InstanceType: "ml.t3.medium", InstanceCount: 1},
			},
			want: 41.61,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSageMakerEndpointMonthlyCost(tt.variants)
			assert.InDelta(t, tt.want, got, 1e-6)
		})
	}
}
