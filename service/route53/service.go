package route53

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/elC0mpa/aws-doctor/model"
)

func NewService(awsconfig aws.Config) *service {
	client := route53.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

// GetEmptyHostedZones returns hosted zones that have only NS and SOA records
// (effectively empty and potentially unused)
func (s *service) GetEmptyHostedZones(ctx context.Context) ([]model.HostedZoneWasteInfo, error) {
	var results []model.HostedZoneWasteInfo

	paginator := route53.NewListHostedZonesPaginator(s.client, &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list hosted zones: %w", err)
		}

		for _, zone := range page.HostedZones {
			// Route 53 hosted zones always have NS and SOA records
			// If ResourceRecordSetCount <= 2, the zone is effectively empty
			recordCount := int64(0)
			if zone.ResourceRecordSetCount != nil {
				recordCount = *zone.ResourceRecordSetCount
			}

			// Consider empty if only NS and SOA (count <= 2)
			isEmpty := recordCount <= 2

			// Extract zone ID without /hostedzone/ prefix
			zoneId := aws.ToString(zone.Id)
			if strings.HasPrefix(zoneId, "/hostedzone/") {
				zoneId = strings.TrimPrefix(zoneId, "/hostedzone/")
			}

			comment := ""
			if zone.Config != nil && zone.Config.Comment != nil {
				comment = *zone.Config.Comment
			}

			isPrivate := false
			if zone.Config != nil {
				isPrivate = zone.Config.PrivateZone
			}

			// Route 53 pricing: $0.50 per hosted zone per month
			monthlyCost := 0.50

			// Only include zones that are empty or have very few records
			if isEmpty {
				results = append(results, model.HostedZoneWasteInfo{
					HostedZoneId:   zoneId,
					Name:           aws.ToString(zone.Name),
					RecordSetCount: recordCount,
					IsPrivate:      isPrivate,
					Comment:        comment,
					MonthlyCost:    monthlyCost,
					IsEmpty:        isEmpty,
				})
			}
		}
	}

	return results, nil
}
