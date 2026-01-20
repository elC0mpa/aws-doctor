package elb

import (
	"context"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type service struct {
	client *elb.Client
}

type ELBService interface {
	GetUnusedLoadBalancers(ctx context.Context) ([]types.LoadBalancer, error)
}
