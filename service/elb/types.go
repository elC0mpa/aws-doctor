package elb

import (
	"context"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type service struct {
	client *elb.Client
}

// Service defines the interface for AWS ELB service.
type Service interface {
	GetUnusedLoadBalancers(ctx context.Context) ([]types.LoadBalancer, error)
}
