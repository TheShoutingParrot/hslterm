package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rivo/tview"
)

func printAlerts(alerts []Alert) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Alert", "Severity", "Date", "Effect", "Link"})

	t.SetStyle(table.StyleRounded)

	fmt.Println(bold("HSL Alerts"))

	for _, alert := range alerts {
		link := hyperlink("Link", alert.AlertUrl)
		if alert.AlertUrl == "" {
			link = "No link"
		}

		t.AppendRow(table.Row{
			alert.AlertHeaderText + "\n",
			alert.AlertSeverityLevel,
			fmt.Sprintf("%v-%v",
				time.Unix(alert.EffectiveStartDate, 0).Format("15:04 01.02"),
				time.Unix(alert.EffectiveEndDate, 0).Format("15:04 01.02")),
			alert.AlertEffect,
			link,
		})
	}

	cols, err := getTerminalWidth()
	if err != nil {
		fmt.Println(redText("failed to get terminal width: " + err.Error()))
	}

	fmt.Println("cols: ", cols)

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: cols >> 1}, // Limit "Location" column to max 30 characters
	})

	t.Render()

}
func tuiDisplayAlerts(alerts []Alert) {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0)

	headers := []string{"Alert", "Severity", "Date", "Effect", "Link"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	for i, alert := range alerts {
		link := "No link"
		if alert.AlertUrl != "" {
			link = alert.AlertUrl
		}

		table.SetCell(i+1, 0, tview.NewTableCell(alert.AlertHeaderText).
			SetTextColor(tview.Styles.PrimaryTextColor))
		table.SetCell(i+1, 1, tview.NewTableCell(alert.AlertSeverityLevel).
			SetTextColor(tview.Styles.PrimaryTextColor))
		table.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%v-%v",
			time.Unix(alert.EffectiveStartDate, 0).Format("15:04 01.02"),
			time.Unix(alert.EffectiveEndDate, 0).Format("15:04 01.02"))).
			SetTextColor(tview.Styles.PrimaryTextColor))
		table.SetCell(i+1, 3, tview.NewTableCell(alert.AlertEffect).
			SetTextColor(tview.Styles.PrimaryTextColor))
		table.SetCell(i+1, 4, tview.NewTableCell(link).
			SetTextColor(tview.Styles.PrimaryTextColor))
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
		case tcell.KeyEsc:
			app.Stop()
		}

		return event
	})

	if err := app.SetRoot(table, true).Run(); err != nil {
		panic(err)
	}
}
