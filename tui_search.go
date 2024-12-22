package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func tuiDisplaySearch(apikey string) {
	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
		case tcell.KeyEsc:
			app.Stop()
		}

		return event
	})

	inputField := tview.NewInputField()

	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewForm().
			AddFormItem(inputField).
			AddButton("Search for stops", func() {
				stops, err := getStopData(apikey, inputField.GetText(), 5)
				if err != nil {
					app.Stop()
					panic(err)
				}

				app.Stop()
				tuiDisplayStops(stops, apikey)
			}), 0, 1, true).
		AddItem(nil, 0, 1, false)

	frame := tview.NewFrame(flex).
		SetBorders(1, 1, 1, 1, 1, 1).
		AddText("Search for HSL stops", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("hslterm", false, tview.AlignCenter, tcell.ColorLightBlue)

	layout := tview.NewFlex().AddItem(frame, 0, 1, true)

	app.SetRoot(layout, true)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}
