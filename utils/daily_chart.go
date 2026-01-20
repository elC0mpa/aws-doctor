package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// DailyJSON represents the JSON output for daily cost analysis
type DailyJSON struct {
	AccountID       string              `json:"account_id"`
	GeneratedAt     string              `json:"generated_at"`
	Days            []DayCostJSON       `json:"days"`
	DayOfWeekStats  []DayOfWeekStatJSON `json:"day_of_week_stats"`
	TotalCost       float64             `json:"total_cost"`
	AverageDailyCost float64            `json:"average_daily_cost"`
}

// DayCostJSON represents cost data for a single day
type DayCostJSON struct {
	Date      string  `json:"date"`
	DayOfWeek string  `json:"day_of_week"`
	Amount    float64 `json:"amount"`
	Unit      string  `json:"unit"`
}

// DayOfWeekStatJSON represents aggregated stats for a day of week
type DayOfWeekStatJSON struct {
	DayOfWeek    string  `json:"day_of_week"`
	TotalCost    float64 `json:"total_cost"`
	AverageCost  float64 `json:"average_cost"`
	Count        int     `json:"count"`
}

// DrawDailyChart displays daily cost data with day-of-week analysis
func DrawDailyChart(accountId string, dailyCosts []model.DailyCostInfo) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 📅 DAILY COST ANALYSIS"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(accountId))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))

	if len(dailyCosts) == 0 {
		fmt.Println("\n" + text.FgHiYellow.Sprint(" No daily cost data available."))
		return
	}

	// Draw daily costs table
	drawDailyCostsTable(dailyCosts)

	// Draw day-of-week analysis
	drawDayOfWeekTable(dailyCosts)
}

func drawDailyCostsTable(dailyCosts []model.DailyCostInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Last 30 Days Cost")

	t.AppendHeader(table.Row{"Date", "Day", "Cost", "Visual"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignLeft},
	})

	// Find max cost for visual scaling
	var maxCost float64
	var totalCost float64
	for _, day := range dailyCosts {
		if day.Amount > maxCost {
			maxCost = day.Amount
		}
		totalCost += day.Amount
	}

	// Display last 14 days to keep it readable
	start := 0
	if len(dailyCosts) > 14 {
		start = len(dailyCosts) - 14
	}

	for _, day := range dailyCosts[start:] {
		// Create visual bar
		barLen := int((day.Amount / maxCost) * 20)
		if barLen < 1 && day.Amount > 0 {
			barLen = 1
		}
		bar := ""
		for i := 0; i < barLen; i++ {
			bar += "█"
		}

		// Color based on day of week
		dayColor := text.FgWhite
		if day.DayOfWeek == "Saturday" || day.DayOfWeek == "Sunday" {
			dayColor = text.FgHiCyan
		}

		t.AppendRow(table.Row{
			day.Date,
			dayColor.Sprint(day.DayOfWeek[:3]),
			fmt.Sprintf("$%.2f", day.Amount),
			text.FgHiGreen.Sprint(bar),
		})
	}

	t.AppendSeparator()
	avgCost := totalCost / float64(len(dailyCosts))
	t.AppendRow(table.Row{
		"",
		text.FgHiWhite.Sprint("Total"),
		text.FgHiWhite.Sprintf("$%.2f", totalCost),
		"",
	})
	t.AppendRow(table.Row{
		"",
		text.FgHiWhite.Sprint("Avg/Day"),
		text.FgHiWhite.Sprintf("$%.2f", avgCost),
		"",
	})

	t.Render()
	fmt.Println()
}

func drawDayOfWeekTable(dailyCosts []model.DailyCostInfo) {
	// Aggregate by day of week
	dowStats := make(map[string]struct {
		total float64
		count int
	})

	for _, day := range dailyCosts {
		stat := dowStats[day.DayOfWeek]
		stat.total += day.Amount
		stat.count++
		dowStats[day.DayOfWeek] = stat
	}

	// Order days properly
	dayOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Day-of-Week Pattern")

	t.AppendHeader(table.Row{"Day", "Avg Cost", "Visual", "Pattern"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
	})

	// Find max average for scaling
	var maxAvg float64
	for _, stat := range dowStats {
		avg := stat.total / float64(stat.count)
		if avg > maxAvg {
			maxAvg = avg
		}
	}

	// Calculate overall average
	var totalAll float64
	var countAll int
	for _, stat := range dowStats {
		totalAll += stat.total
		countAll += stat.count
	}
	overallAvg := totalAll / float64(countAll)

	for _, day := range dayOrder {
		stat, ok := dowStats[day]
		if !ok {
			continue
		}

		avg := stat.total / float64(stat.count)

		// Create visual bar
		barLen := int((avg / maxAvg) * 15)
		if barLen < 1 && avg > 0 {
			barLen = 1
		}
		bar := ""
		for i := 0; i < barLen; i++ {
			bar += "█"
		}

		// Determine pattern indicator
		pattern := ""
		diff := ((avg - overallAvg) / overallAvg) * 100
		if diff > 20 {
			pattern = text.FgHiRed.Sprintf("↑ +%.0f%%", diff)
		} else if diff < -20 {
			pattern = text.FgHiGreen.Sprintf("↓ %.0f%%", diff)
		} else {
			pattern = text.FgHiWhite.Sprint("≈ avg")
		}

		// Color weekends differently
		dayColor := text.FgWhite
		barColor := text.FgHiBlue
		if day == "Saturday" || day == "Sunday" {
			dayColor = text.FgHiCyan
			barColor = text.FgHiCyan
		}

		t.AppendRow(table.Row{
			dayColor.Sprint(day),
			fmt.Sprintf("$%.2f", avg),
			barColor.Sprint(bar),
			pattern,
		})
	}

	t.Render()
	fmt.Println()

	// Print insights
	printDailyInsights(dowStats, overallAvg)
}

func printDailyInsights(dowStats map[string]struct {
	total float64
	count int
}, overallAvg float64) {
	fmt.Println(text.FgHiWhite.Sprint(" 💡 Insights:"))

	// Find highest and lowest days
	var highestDay, lowestDay string
	var highestAvg, lowestAvg float64 = 0, 999999999

	for day, stat := range dowStats {
		avg := stat.total / float64(stat.count)
		if avg > highestAvg {
			highestAvg = avg
			highestDay = day
		}
		if avg < lowestAvg {
			lowestAvg = avg
			lowestDay = day
		}
	}

	fmt.Printf("    • Highest spend: %s ($%.2f avg)\n", highestDay, highestAvg)
	fmt.Printf("    • Lowest spend: %s ($%.2f avg)\n", lowestDay, lowestAvg)

	// Weekend vs weekday comparison
	var weekdayTotal, weekendTotal float64
	var weekdayCount, weekendCount int
	for day, stat := range dowStats {
		if day == "Saturday" || day == "Sunday" {
			weekendTotal += stat.total
			weekendCount += stat.count
		} else {
			weekdayTotal += stat.total
			weekdayCount += stat.count
		}
	}

	if weekdayCount > 0 && weekendCount > 0 {
		weekdayAvg := weekdayTotal / float64(weekdayCount)
		weekendAvg := weekendTotal / float64(weekendCount)
		diff := ((weekendAvg - weekdayAvg) / weekdayAvg) * 100

		if diff > 10 {
			fmt.Printf("    • Weekend spending is %.0f%% higher than weekdays\n", diff)
		} else if diff < -10 {
			fmt.Printf("    • Weekend spending is %.0f%% lower than weekdays\n", -diff)
		} else {
			fmt.Println("    • Weekend and weekday spending are similar")
		}
	}

	fmt.Println()
}

// OutputDailyJSON outputs daily cost data as JSON
func OutputDailyJSON(accountID string, dailyCosts []model.DailyCostInfo) error {
	output := DailyJSON{
		AccountID:   accountID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Days:        make([]DayCostJSON, 0, len(dailyCosts)),
	}

	// Aggregate by day of week
	dowStats := make(map[string]struct {
		total float64
		count int
	})

	var totalCost float64
	for _, day := range dailyCosts {
		output.Days = append(output.Days, DayCostJSON{
			Date:      day.Date,
			DayOfWeek: day.DayOfWeek,
			Amount:    day.Amount,
			Unit:      day.Unit,
		})

		stat := dowStats[day.DayOfWeek]
		stat.total += day.Amount
		stat.count++
		dowStats[day.DayOfWeek] = stat
		totalCost += day.Amount
	}

	output.TotalCost = totalCost
	if len(dailyCosts) > 0 {
		output.AverageDailyCost = totalCost / float64(len(dailyCosts))
	}

	// Build day-of-week stats
	dayOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range dayOrder {
		if stat, ok := dowStats[day]; ok {
			output.DayOfWeekStats = append(output.DayOfWeekStats, DayOfWeekStatJSON{
				DayOfWeek:   day,
				TotalCost:   stat.total,
				AverageCost: stat.total / float64(stat.count),
				Count:       stat.count,
			})
		}
	}

	// Sort days by date
	sort.Slice(output.Days, func(i, j int) bool {
		return output.Days[i].Date < output.Days[j].Date
	})

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
