package awscostexplorer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

type service struct {
	client *ec2.Client
}

type EC2Service interface {
	GetElasticIpAddressesInfo(ctx context.Context) (*model.ElasticIpInfo, error)
	GetUnusedElasticIpAddressesInfo(ctx context.Context) ([]types.Address, error)
	GetUnusedEBSVolumes(ctx context.Context) ([]types.Volume, error)
	GetStoppedInstancesInfo(ctx context.Context) ([]types.Instance, []types.Volume, error)
	GetReservedInstanceExpiringOrExpired30DaysWaste(ctx context.Context) ([]model.RiExpirationInfo, error)
	GetUnusedAMIs(ctx context.Context, staleDays int) ([]model.AMIWasteInfo, error)
}
