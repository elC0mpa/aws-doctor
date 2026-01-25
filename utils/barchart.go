package utils //nolint:revive

import (
	"fmt"
	"sort"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/text"
)

// ColorRank1 is the color for the highest ranked cost.
const (
	ColorRank1 = "#d73027"
	ColorRank2 = "#f46d43"
	ColorRank3 = "#fee08b"
	ColorRank4 = "#abdda4"
	ColorRank5 = "#66c2a5"
	ColorRank6 = "#1a9850"
)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#F4D060"))

// DrawTrendChart draws a bar chart of monthly costs.
func DrawTrendChart(accountID string, monthlyCosts []model.CostInfo) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 📈 AWS DOCTOR TREND"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(accountID))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))

	bc := barchart.New(130, 20)

	indexedColors := assignRankedColors(monthlyCosts)

	for idx, monthlyCost := range monthlyCosts {
		data := barchart.BarData{
			Label: getBarLabel(*monthlyCost.Start, monthlyCost),
			Values: []barchart.BarValue{
				{
					Value: monthlyCost.CostGroup["Total"].Amount,
					Style: lipgloss.NewStyle().Foreground(lipgloss.Color(indexedColors[idx])),
				},
			},
		}

		bc.Push(data)
	}

	fmt.Println()
	fmt.Println()

	bc.Draw()
	s := lipgloss.JoinHorizontal(lipgloss.Top,
		defaultStyle.Render(bc.View()),
	)

	fmt.Println(s)
}

func getBarLabel(date string, monthlyCost model.CostInfo) string {
	parsedTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Sprintf("%s: %.2f %s", date, monthlyCost.CostGroup["Total"].Amount, monthlyCost.CostGroup["Total"].Unit)
	}

	return fmt.Sprintf("%s: %.2f %s", parsedTime.Format("Jan"), monthlyCost.CostGroup["Total"].Amount, monthlyCost.CostGroup["Total"].Unit)
}

func assignRankedColors(allCosts []model.CostInfo) []string {
	palette := []string{ColorRank1, ColorRank2, ColorRank3, ColorRank4, ColorRank5, ColorRank6}

	type costWithIndex struct {
		index int
		value float64
	}

	costsToSort := make([]costWithIndex, len(allCosts))
	for i, cost := range allCosts {
		costsToSort[i] = costWithIndex{
			index: i,
			value: cost.CostGroup["Total"].Amount,
		}
	}

	sort.Slice(costsToSort, func(i, j int) bool {
		return costsToSort[i].value > costsToSort[j].value
	})

	resultColors := make([]string, len(allCosts))

	for rank, sortedCost := range costsToSort {
		originalIndex := sortedCost.index
		if rank < len(palette) {
			resultColors[originalIndex] = palette[rank]
		}
	}

	return resultColors
}
