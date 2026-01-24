package awscostexplorer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/elC0mpa/aws-doctor/model"
)

const (
	unblendedCost = "UnblendedCost"
)

func NewService(awsconfig aws.Config) *service {
	client := costexplorer.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

func (s *service) GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	return s.GetMonthCostsByService(ctx, time.Now())
}

func (s *service) GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	return s.GetMonthCostsByService(ctx, oneMonthAgo)
}

func (s *service) GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error) {
	firstOfMonth := s.getFirstDayOfMonth(endDate)
	firstOfMonthStr := firstOfMonth.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(firstOfMonthStr),
			End:   aws.String(endDate.Format("2006-01-02")),
		},
		Metrics: []string{unblendedCost},
		GroupBy: []types.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: types.GroupDefinitionTypeDimension,
			},
		},
	}

	output, err := s.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, err
	}

	return &model.CostInfo{
		CostGroup:    s.filterGroups(output.ResultsByTime[0].Groups, unblendedCost),
		DateInterval: *output.ResultsByTime[0].TimePeriod,
	}, nil
}

func (s *service) GetCurrentMonthTotalCosts(ctx context.Context) (*string, error) {
	return s.GetMonthTotalCosts(ctx, time.Now())
}

func (s *service) GetLastMonthTotalCosts(ctx context.Context) (*string, error) {
	return s.GetMonthTotalCosts(ctx, time.Now().AddDate(0, -1, 0))
}

func (s *service) GetLastSixMonthsCosts(ctx context.Context) ([]model.CostInfo, error) {
	firstOfMonth := s.getFirstDayOfMonth(time.Now().AddDate(0, -6, 0))
	firstOfMonthStr := firstOfMonth.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(firstOfMonthStr),
			End:   aws.String(s.getFirstDayOfMonth(time.Now()).Format("2006-01-02")),
		},
		Metrics: []string{unblendedCost},
	}

	output, err := s.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, err
	}

	monthlyCosts := make([]model.CostInfo, 0, len(output.ResultsByTime))

	for _, timeResult := range output.ResultsByTime {
		amount, _ := strconv.ParseFloat(*timeResult.Total[unblendedCost].Amount, 64)
		costGroups := make(map[string]struct {
			Amount float64
			Unit   string
		})

		costGroups["Total"] = struct {
			Amount float64
			Unit   string
		}{
			Amount: amount,
			Unit:   *timeResult.Total[unblendedCost].Unit,
		}

		monthlyCost := model.CostInfo{
			DateInterval: *timeResult.TimePeriod,
			CostGroup:    costGroups,
		}
		monthlyCosts = append(monthlyCosts, monthlyCost)
	}

	return monthlyCosts, nil
}

func (s *service) GetMonthTotalCosts(ctx context.Context, endDate time.Time) (*string, error) {
	firstOfMonth := s.getFirstDayOfMonth(endDate)
	firstOfMonthStr := firstOfMonth.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(firstOfMonthStr),
			End:   aws.String(endDate.Format("2006-01-02")),
		},
		Metrics: []string{unblendedCost},
	}

	output, err := s.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.ResultsByTime) == 0 {
		return nil, fmt.Errorf("no cost data returned for the specified time period")
	}

	totalInfo, ok := output.ResultsByTime[0].Total[unblendedCost]
	if !ok || totalInfo.Amount == nil {
		return nil, fmt.Errorf("cost data missing %s metric", unblendedCost)
	}

	amount, err := strconv.ParseFloat(*totalInfo.Amount, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse total amount %q: %w", *totalInfo.Amount, err)
	}

	total := fmt.Sprintf("%.2f %s", amount, *totalInfo.Unit)
	return &total, nil
}

func (s *service) getFirstDayOfMonth(month time.Time) time.Time {
	return time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
}

func (s *service) getLastDayOfMonth(month time.Time) time.Time {
	return time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location())
}

func (s *service) filterGroups(results []types.Group, costsAggregation string) model.CostGroup {
	filtered := make([]types.Group, 0, len(results))

	for _, g := range results {
		amountStr := ""
		if metric, ok := g.Metrics[costsAggregation]; ok && metric.Amount != nil {
			amountStr = *metric.Amount
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount == 0 {
			continue
		}
		filtered = append(filtered, g)
	}

	costGroups := make(map[string]struct {
		Amount float64
		Unit   string
	})

	for _, v := range filtered {
		amount, _ := strconv.ParseFloat(*v.Metrics[costsAggregation].Amount, 64)
		costGroups[v.Keys[0]] = struct {
			Amount float64
			Unit   string
		}{
			Amount: amount,
			Unit:   *v.Metrics[costsAggregation].Unit,
		}
	}

	return costGroups
}
