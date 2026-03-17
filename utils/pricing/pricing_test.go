package pricing

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestEBSCostPerGBMonth(t *testing.T) {
	tests := []struct {
		name       string
		volumeType types.VolumeType
		want       float64
	}{
		{
			name:       "gp2 volume",
			volumeType: types.VolumeTypeGp2,
			want:       EBSgp2CostPerGBMonth,
		},
		{
			name:       "gp3 volume",
			volumeType: types.VolumeTypeGp3,
			want:       EBSgp3CostPerGBMonth,
		},
		{
			name:       "io1 volume",
			volumeType: types.VolumeTypeIo1,
			want:       EBSio1CostPerGBMonth,
		},
		{
			name:       "io2 volume",
			volumeType: types.VolumeTypeIo2,
			want:       EBSio2CostPerGBMonth,
		},
		{
			name:       "st1 volume",
			volumeType: types.VolumeTypeSt1,
			want:       EBSst1CostPerGBMonth,
		},
		{
			name:       "sc1 volume",
			volumeType: types.VolumeTypeSc1,
			want:       EBSsc1CostPerGBMonth,
		},
		{
			name:       "unknown type defaults to gp2",
			volumeType: types.VolumeType("unknown"),
			want:       EBSgp2CostPerGBMonth,
		},
		{
			name:       "empty type defaults to gp2",
			volumeType: types.VolumeType(""),
			want:       EBSgp2CostPerGBMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EBSCostPerGBMonth(tt.volumeType)
			assert.Equal(t, tt.want, got)
		})
	}
}
