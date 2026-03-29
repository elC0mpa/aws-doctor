// Package awscostexplorer provides a service for interacting with AWS Cost Explorer.
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

// NewService creates a new Cost Explorer service.
func NewService(awsconfig aws.Config) Service {
	client := costexplorer.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

func (s *service) GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	now := time.Now()
	// AWS TimePeriod End is exclusive. To include today, we must use tomorrow in the API.
	tomorrow := now.AddDate(0, 0, 1)

	costInfo, err := s.GetMonthCostsByService(ctx, tomorrow)
	if err != nil {
		return nil, err
	}

	// Adjust the metadata so the UI shows today as the end date
	costInfo.End = aws.String(now.Format("2006-01-02"))

	return costInfo, nil
}

func (s *service) GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	now := time.Now()
	firstOfCurrentMonth := s.getFirstDayOfMonth(now)
	firstOfLastMonth := firstOfCurrentMonth.AddDate(0, -1, 0)
	lastOfLastMonth := s.getLastDayOfMonth(firstOfLastMonth)

	day := now.Day()
	if day > lastOfLastMonth.Day() {
		day = lastOfLastMonth.Day()
	}

	// To include 'day', we use 'day + 1' as the exclusive end date for the API
	endOfLastMonthAPI := firstOfLastMonth.AddDate(0, 0, day)

	costInfo, err := s.getCostsByPeriod(ctx, firstOfLastMonth, endOfLastMonthAPI)
	if err != nil {
		return nil, err
	}

	// Adjust metadata so the UI shows the intended day
	displayEnd := firstOfLastMonth.AddDate(0, 0, day-1)
	costInfo.End = aws.String(displayEnd.Format("2006-01-02"))

	return costInfo, nil
}

func (s *service) GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error) {
	if time.Now().Day() == 1 {
		return nil, model.ErrFirstDayOfMonth
	}

	firstOfMonth := s.getFirstDayOfMonth(endDate)

	return s.getCostsByPeriod(ctx, firstOfMonth, endDate)
}

func (s *service) getCostsByPeriod(ctx context.Context, start, end time.Time) (*model.CostInfo, error) {
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	if startStr == endStr {
		return &model.CostInfo{
			CostGroup: make(model.CostGroup),
			DateInterval: types.DateInterval{
				Start: aws.String(startStr),
				End:   aws.String(endStr),
			},
		}, nil
	}

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(startStr),
			End:   aws.String(endStr),
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

	if len(output.ResultsByTime) == 0 {
		return &model.CostInfo{
			CostGroup: make(model.CostGroup),
			DateInterval: types.DateInterval{
				Start: aws.String(startStr),
				End:   aws.String(endStr),
			},
		}, nil
	}

	return &model.CostInfo{
		CostGroup:    s.filterGroups(output.ResultsByTime[0].Groups, unblendedCost),
		DateInterval: *output.ResultsByTime[0].TimePeriod,
	}, nil
}

func (s *service) GetCurrentMonthTotalCosts(ctx context.Context) (*string, error) {
	now := time.Now()
	// AWS TimePeriod End is exclusive. To include today, we must use tomorrow in the API.
	tomorrow := now.AddDate(0, 0, 1)

	return s.getTotalCostsByPeriod(ctx, s.getFirstDayOfMonth(now), tomorrow)
}

func (s *service) GetLastMonthTotalCosts(ctx context.Context) (*string, error) {
	now := time.Now()
	firstOfCurrentMonth := s.getFirstDayOfMonth(now)
	firstOfLastMonth := firstOfCurrentMonth.AddDate(0, -1, 0)
	lastOfLastMonth := s.getLastDayOfMonth(firstOfLastMonth)

	day := now.Day()
	if day > lastOfLastMonth.Day() {
		day = lastOfLastMonth.Day()
	}

	// To include 'day', we use 'day + 1' as the exclusive end date for the API
	endOfLastMonthAPI := firstOfLastMonth.AddDate(0, 0, day)

	return s.getTotalCostsByPeriod(ctx, firstOfLastMonth, endOfLastMonthAPI)
}

func (s *service) GetLastSixMonthsCosts(ctx context.Context, services []string) ([]model.CostInfo, error) {
	now := time.Now()
	firstOfCurrentMonth := s.getFirstDayOfMonth(now)
	firstOfMonth := firstOfCurrentMonth.AddDate(0, -6, 0)
	firstOfMonthStr := firstOfMonth.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(firstOfMonthStr),
			End:   aws.String(firstOfCurrentMonth.Format("2006-01-02")),
		},
		Metrics: []string{unblendedCost},
	}

	if len(services) > 0 {
		input.Filter = &types.Expression{
			Dimensions: &types.DimensionValues{
				Key:    types.Dimension("SERVICE"),
				Values: services,
			},
		}
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

	return s.getTotalCostsByPeriod(ctx, firstOfMonth, endDate)
}

func (s *service) getTotalCostsByPeriod(ctx context.Context, start, end time.Time) (*string, error) {
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	if startStr == endStr {
		zero := "0.00 USD"
		return &zero, nil
	}

	input := &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityMonthly,
		TimePeriod: &types.DateInterval{
			Start: aws.String(startStr),
			End:   aws.String(endStr),
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
