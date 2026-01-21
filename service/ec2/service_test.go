package awscostexplorer

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestGetResourceTypeFromDescription(t *testing.T) {
	// Create a service instance for testing the method
	// We don't need a real client since getResourceTypeFromDescription doesn't use it
	s := &service{}

	tests := []struct {
		name        string
		description string
		want        types.NetworkInterfaceType
	}{
		// Application Load Balancer cases
		{
			name:        "alb_standard_format",
			description: "ELB app/my-load-balancer/abc123",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},
		{
			name:        "alb_lowercase",
			description: "elb app/test-alb/def456",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},
		{
			name:        "alb_mixed_case",
			description: "ELB APP/MyALB/xyz789",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},

		// Network Load Balancer cases
		{
			name:        "nlb_standard_format",
			description: "ELB net/my-nlb/abc123",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
		},
		{
			name:        "nlb_lowercase",
			description: "elb net/test-nlb/def456",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
		},

		// NAT Gateway cases
		{
			name:        "nat_gateway_standard",
			description: "Interface for NAT Gateway nat-0abc123def456",
			want:        types.NetworkInterfaceTypeNatGateway,
		},
		{
			name:        "nat_gateway_hyphenated",
			description: "nat-gateway interface",
			want:        types.NetworkInterfaceTypeNatGateway,
		},
		{
			name:        "nat_gateway_with_id",
			description: "NAT Gateway nat-12345",
			want:        types.NetworkInterfaceTypeNatGateway,
		},

		// Global Accelerator cases
		{
			name:        "global_accelerator",
			description: "AWS GlobalAccelerator managed interface",
			want:        types.NetworkInterfaceTypeGlobalAcceleratorManaged,
		},
		{
			name:        "global_accelerator_lowercase",
			description: "globalaccelerator endpoint",
			want:        types.NetworkInterfaceTypeGlobalAcceleratorManaged,
		},

		// VPC Endpoint cases
		{
			name:        "vpc_endpoint_standard",
			description: "VPC Endpoint Interface vpce-0abc123",
			want:        types.NetworkInterfaceTypeVpcEndpoint,
		},
		{
			name:        "vpc_endpoint_with_id",
			description: "Interface for vpce-12345678",
			want:        types.NetworkInterfaceTypeVpcEndpoint,
		},

		// Transit Gateway cases
		{
			name:        "transit_gateway_standard",
			description: "Transit Gateway Attachment tgw-attach-123",
			want:        types.NetworkInterfaceTypeTransitGateway,
		},
		{
			name:        "transit_gateway_with_id",
			description: "Network interface for tgw-12345",
			want:        types.NetworkInterfaceTypeTransitGateway,
		},

		// Lambda cases
		{
			name:        "lambda_standard",
			description: "AWS Lambda VPC ENI-my-function-abc123",
			want:        types.NetworkInterfaceTypeLambda,
		},
		{
			name:        "lambda_lowercase",
			description: "aws lambda function interface",
			want:        types.NetworkInterfaceTypeLambda,
		},

		// API Gateway cases
		{
			name:        "api_gateway_standard",
			description: "API Gateway managed interface",
			want:        types.NetworkInterfaceTypeApiGatewayManaged,
		},
		{
			name:        "api_gateway_lowercase",
			description: "api gateway endpoint",
			want:        types.NetworkInterfaceTypeApiGatewayManaged,
		},

		// IoT Rules cases
		{
			name:        "iot_rules",
			description: "IoT Rules managed interface",
			want:        types.NetworkInterfaceTypeIotRulesManaged,
		},

		// Gateway Load Balancer cases
		{
			name:        "gwlb_standard",
			description: "Gateway Load Balancer Endpoint",
			want:        types.NetworkInterfaceTypeGatewayLoadBalancer,
		},

		// Custom resource types (returned as NetworkInterfaceType strings)
		{
			name:        "redshift_cluster",
			description: "Redshift cluster my-cluster",
			want:        types.NetworkInterfaceType("redshift_cluster"),
		},
		{
			name:        "rds_database",
			description: "RDS database instance",
			want:        types.NetworkInterfaceType("rds_database"),
		},
		{
			name:        "directory_service",
			description: "Directory Service interface",
			want:        types.NetworkInterfaceType("directory_service"),
		},
		{
			name:        "fsx_filesystem",
			description: "FSx file system interface",
			want:        types.NetworkInterfaceType("fsx"),
		},

		// Default/fallback cases
		{
			name:        "empty_description",
			description: "",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "unknown_description",
			description: "Some random network interface",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "ec2_instance_description",
			description: "Primary network interface",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "ecs_task_description",
			description: "ecs-task/12345",
			want:        types.NetworkInterfaceType("interface"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getResourceTypeFromDescription(tt.description)
			if got != tt.want {
				t.Errorf("getResourceTypeFromDescription(%q) = %v, want %v", tt.description, got, tt.want)
			}
		})
	}
}

func TestGetResourceTypeFromDescription_CaseInsensitivity(t *testing.T) {
	s := &service{}

	// Test that matching is case-insensitive
	casePairs := []struct {
		lower string
		upper string
		mixed string
	}{
		{"elb app/test/123", "ELB APP/TEST/123", "Elb App/Test/123"},
		{"nat gateway", "NAT GATEWAY", "Nat Gateway"},
		{"aws lambda", "AWS LAMBDA", "Aws Lambda"},
		{"vpc endpoint", "VPC ENDPOINT", "Vpc Endpoint"},
	}

	for _, pair := range casePairs {
		lowerResult := s.getResourceTypeFromDescription(pair.lower)
		upperResult := s.getResourceTypeFromDescription(pair.upper)
		mixedResult := s.getResourceTypeFromDescription(pair.mixed)

		if lowerResult != upperResult || upperResult != mixedResult {
			t.Errorf("Case sensitivity issue: lower=%v, upper=%v, mixed=%v for inputs %q/%q/%q",
				lowerResult, upperResult, mixedResult, pair.lower, pair.upper, pair.mixed)
		}
	}
}

func TestGetResourceTypeFromDescription_Priority(t *testing.T) {
	s := &service{}

	// Test that when multiple keywords could match, the first matching condition wins
	// Based on the order in the implementation
	tests := []struct {
		name        string
		description string
		want        types.NetworkInterfaceType
		reason      string
	}{
		{
			name:        "alb_before_generic_elb",
			description: "ELB app/lb-name/123 some elb",
			want:        types.NetworkInterfaceTypeLoadBalancer,
			reason:      "app/ should match before any other ELB pattern",
		},
		{
			name:        "nlb_before_generic_elb",
			description: "ELB net/lb-name/123",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
			reason:      "net/ should identify NLB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getResourceTypeFromDescription(tt.description)
			if got != tt.want {
				t.Errorf("getResourceTypeFromDescription(%q) = %v, want %v (%s)",
					tt.description, got, tt.want, tt.reason)
			}
		})
	}
}

func BenchmarkGetResourceTypeFromDescription(b *testing.B) {
	s := &service{}
	descriptions := []string{
		"ELB app/my-load-balancer/abc123",
		"Interface for NAT Gateway nat-0abc123def456",
		"AWS Lambda VPC ENI-my-function-abc123",
		"Primary network interface",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, desc := range descriptions {
			s.getResourceTypeFromDescription(desc)
		}
	}
}
