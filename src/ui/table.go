package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/denwerk/moco/src/types"
)

// German day names
var dayTranslations = map[string]string{
	"Monday":    "Montag",
	"Tuesday":   "Dienstag",
	"Wednesday": "Mittwoch",
	"Thursday":  "Donnerstag",
	"Friday":    "Freitag",
	"Saturday":  "Samstag",
	"Sunday":    "Sonntag",
}

// CreateTimeEntriesTable creates a new table with the given time entries
func CreateTimeEntriesTable(entries []types.TimeEntry, height int) table.Model {
	// Group entries by date
	entriesByDate := make(map[string][]types.TimeEntry)
	for _, entry := range entries {
		date := entry.Date
		entriesByDate[date] = append(entriesByDate[date], entry)
	}

	// Create table columns
	columns := []table.Column{
		{Title: "Entry", Width: 30},
		{Title: "Hours", Width: 12},
		{Title: "Task", Width: 40},
	}

	// Create rows with date headers
	var rows []table.Row

	// Get sorted dates (latest first)
	var dates []string
	for date := range entriesByDate {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// Add rows in date order
	for _, date := range dates {
		entries := entriesByDate[date]
		// Parse date and format in German style with day of week
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			// If parsing fails, use the original date
			rows = append(rows, table.Row{
				TitleStyle.Render(date),
				"",
				"",
			})
		} else {
			// Format date in German style: "Monday, 02.01.2006"
			formattedDate := parsedDate.Format("Monday, 02.01.2006")
			// Translate day name to German
			parts := strings.Split(formattedDate, ", ")
			if len(parts) == 2 {
				day := parts[0]
				if germanDay, ok := dayTranslations[day]; ok {
					formattedDate = germanDay + ", " + parts[1]
				}
			}
			rows = append(rows, table.Row{
				HeaderStyle.Render(formattedDate),
				"",
				"",
			})
		}

		// Add entries for this date
		for _, entry := range entries {
			rows = append(rows, table.Row{
				entry.Description,
				fmt.Sprintf("%.2f", entry.Hours),
				entry.Task.Name,
			})
		}

		// Add total hours for the day
		totalHours := 0.0
		for _, entry := range entries {
			totalHours += entry.Hours
		}
		rows = append(rows, table.Row{
			TotalStyle.Render("Total:"),
			TotalStyle.Render(fmt.Sprintf("%.2f", totalHours)),
			"",
		})

		// Add separator row
		rows = append(rows, table.Row{
			"", "", "",
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	// Customize table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderBottom(true).
		BorderBottomForeground(lipgloss.Color("63")).
		Foreground(lipgloss.Color("255"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("255"))

	t.SetStyles(s)
	return t
}
