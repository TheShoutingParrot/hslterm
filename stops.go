package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rivo/tview"
)

func printStop(stop Stop) {
	fmt.Printf(bold("Stop: %v (%v, %v) %v")+"\n", stop.Name, stop.Desc, stop.Code,
		transportModeEmoji(stop.VehicleMode))
	fmt.Printf(bold("Location: %v, %v")+"\n", stop.Lat, stop.Lon)

	fmt.Println("Routes: ")
	for _, route := range stop.Routes {
		fmt.Printf("%v\t%v - %v\n", transportModeEmoji(route.Mode), route.ShortName, route.LongName)
	}
	fmt.Print("\n")

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Route", "Departing", "Time left"})

	t.SetStyle(table.StyleRounded)

	for _, stopTime := range stop.StopTimes {
		stopTime.RealtimeDeparture += stopTime.ServiceDay

		routeName := fmt.Sprintf("%v - %v", bold(stopTime.Trip.RouteShortName), stopTime.Headsign)
		if stopTime.RealtimeState == "CANCELED" {
			routeName = redText(routeName + " (CANCELED)")
		} else if stopTime.Headsign == "" {
			routeName = bold(stopTime.Trip.RouteShortName)
		}

		tim := formatTimeLeft(stopTime.RealtimeDeparture)
		if tim == "Now" {
			tim = bold(tim)
		}

		t.AppendRow(table.Row{
			routeName,
			time.Unix(stopTime.RealtimeDeparture, 0).Format("15:04"),
			tim,
		})
	}

	// print alerts if there are any
	if len(stop.Alerts) > 0 {
		fmt.Println(redText("\nAlerts:"))
		for _, alert := range stop.Alerts {
			fmt.Printf(redText("\t%v: %v\n"), alert.AlertSeverityLevel, alert.AlertHeaderText)
		}
		fmt.Print("\n")
	}

	t.Render()
}

func newTuiStopFrame(stop Stop, left string, right string) (tview.Primitive, error) {
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 3).
		SetCell(0, 0, tview.NewTableCell("Route").SetAlign(tview.AlignCenter).SetExpansion(1)).
		SetCell(0, 1, tview.NewTableCell("Departing").SetAlign(tview.AlignCenter).SetExpansion(1)).
		SetCell(0, 2, tview.NewTableCell("Time left").SetExpansion(1))

	for i, stopTime := range stop.StopTimes {
		stopTime.RealtimeDeparture += stopTime.ServiceDay

		routeName := fmt.Sprintf("[white]%v - %v", stopTime.Trip.RouteShortName, stopTime.Headsign)
		if stopTime.RealtimeState == "CANCELED" {
			routeName = "[bold][red]" + routeName + " (CANCELED)[-]"
		} else if stopTime.Headsign == "" {
			routeName = "[bold]" + stopTime.Trip.RouteShortName + "[-]"
		}

		table.SetCellSimple(i+1, 0, routeName)
		table.SetCellSimple(i+1, 1, time.Unix(stopTime.RealtimeDeparture, 0).Format("15:04"))
		table.SetCellSimple(i+1, 2, formatTimeLeft(stopTime.RealtimeDeparture))
	}

	routesText := "Routes:"
	for _, route := range stop.Routes {
		routesText += fmt.Sprintf(" %v %v,",
			transportModeEmoji(route.Mode),
			route.ShortName,
		)
	}
	routesText = routesText[:len(routesText)-1]

	frame := tview.NewFrame(table).
		SetBorders(1, 1, 2, 2, 4, 4).
		AddText("hslterm", true, tview.AlignLeft, tcell.ColorWhite).
		AddText(time.Now().Format("15:04 02.01.2006"), true, tview.AlignRight, tcell.ColorWhite).
		AddText(
			fmt.Sprintf("%v %v %v",
				transportModeEmoji(stop.VehicleMode),
				stop.Name,
				transportModeEmoji(stop.VehicleMode)),
			true, tview.AlignCenter, tcell.ColorWhite).
		AddText(fmt.Sprintf("%v (%v)", stop.Desc, stop.Code), true, tview.AlignCenter, tcell.ColorRed).
		AddText("m for menu", false, tview.AlignCenter, tcell.ColorLightBlue).
		AddText(routesText, false, tview.AlignCenter, tcell.ColorBlue)

	if left != "" {
		frame.AddText("← ("+left+")", false, tview.AlignLeft, tcell.ColorBlue)
	}
	if right != "" {
		frame.AddText("("+right+") →", false, tview.AlignRight, tcell.ColorBlue)
	}

	return frame, nil
}

func stopsGetLeftRight(stops []Stop, i int) (left string, right string) {
	if i == 0 {
		left = fmt.Sprintf("%v - %v", stops[len(stops)-1].Code, stops[len(stops)-1].Desc)
		right = fmt.Sprintf("%v - %v", stops[i+1].Desc, stops[i+1].Code)
	} else if i == len(stops)-1 {
		left = fmt.Sprintf("%v - %v", stops[i-1].Code, stops[i-1].Desc)
		right = fmt.Sprintf("%v - %v", stops[0].Code, stops[0].Code)
	} else {
		left = fmt.Sprintf("%v - %v", stops[i-1].Code, stops[i-1].Desc)
		right = fmt.Sprintf("%v - %v", stops[i+1].Code, stops[i+1].Desc)
	}

	return
}

func tuiDisplayStops(stops []Stop, apikey string) {
	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				app.Stop()
			}
		case tcell.KeyCtrlC:
			app.Stop()

		case tcell.KeyEsc:
			app.Stop()
			tuiDisplaySearch(apikey)
		}
		return event
	})

	var frame tview.Primitive
	var err error

	if len(stops) == 0 {
		fmt.Println("no stops found")
		os.Exit(0)
	} else if len(stops) == 1 {
		frame, err = newTuiStopFrame(stops[0], "", "")
		if err != nil {
			os.Exit(1)
		}

		if err := app.SetRoot(frame, true).Run(); err != nil {
			panic(err)
		}
	} else {
		frame, err = newTuiStopFrame(
			stops[0],
			fmt.Sprintf("%v - %v", stops[len(stops)-1].Code, stops[len(stops)-1].Desc),
			fmt.Sprintf("%v - %v", stops[1].Desc, stops[1].Code),
		)
		layout := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(frame, 0, 1, true)

		onMenu := false
		i := 0

		layout.SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyRune:
					if event.Rune() == 'm' {
						if onMenu {
							layout.Clear()
							left, right := stopsGetLeftRight(stops, i)
							newframe, err := newTuiStopFrame(stops[i], left, right)
							if err != nil {
								panic(err)
							}

							layout.Clear()
							layout.AddItem(newframe, 0, 1, true)
							app.SetRoot(layout, true)

							onMenu = false
						} else {
							list := tview.NewList()
							list.AddItem("Cancel", "", 'c', func() {
								left, right := stopsGetLeftRight(stops, i)
								newframe, err := newTuiStopFrame(stops[i], left, right)
								if err != nil {
									panic(err)
								}

								layout.Clear()
								layout.AddItem(newframe, 0, 1, true)
								app.SetRoot(layout, true)

								onMenu = false
							})

							for stopIndex, stop := range stops {
								buttonTitle := fmt.Sprintf("%v %v - %v", stop.Name, stop.Code, stop.Desc)

								shortcut := rune(0)

								if stopIndex < 9 {
									shortcut = rune('1' + stopIndex)
								} else if stopIndex == 9 {
									shortcut = '0'
								}

								list.AddItem(buttonTitle, "", shortcut, func() {
									left, right := stopsGetLeftRight(stops, stopIndex)
									newframe, err := newTuiStopFrame(stop, left, right)
									if err != nil {
										panic(err)
									}

									layout.Clear()
									layout.AddItem(newframe, 0, 1, true)
									app.SetRoot(layout, true)

									onMenu = false
								})
							}

							list.AddItem("Quit", "", 0, func() {
								app.Stop()
							})

							layout.Clear()
							frame := tview.NewFrame(list).
								SetBorders(1, 1, 1, 1, 2, 2).
								AddText("Select the stop to view", true, tview.AlignCenter, tcell.ColorWhite).
								AddText("hslterm", false, tview.AlignCenter, tcell.ColorLightBlue)
							layout.Clear().AddItem(frame, 0, 1, true)
							app.SetRoot(layout, true)
							onMenu = true
						}
					}

				case tcell.KeyLeft:
					if onMenu {
						break
					}

					if i == 0 {
						i = len(stops) - 1
					} else {
						i--
					}

					left, right := stopsGetLeftRight(stops, i)

					newframe, err := newTuiStopFrame(stops[i], left, right)

					if err != nil {
						panic(err)
					}

					layout.Clear()
					layout.AddItem(newframe, 0, 1, true)
					app.SetRoot(layout, true)
				case tcell.KeyRight:
					if onMenu {
						break
					}

					if i == len(stops)-1 {
						i = 0
					} else {
						i++
					}

					left, right := stopsGetLeftRight(stops, i)

					newframe, err := newTuiStopFrame(stops[i], left, right)

					if err != nil {
						panic(err)
					}

					layout.Clear()
					layout.AddItem(newframe, 0, 1, true)
					app.SetRoot(layout, true)
				}

				return event
			},
		)

		if err != nil {
			os.Exit(1)
		}

		go func() {
			ticker := time.NewTicker(20 * time.Second)
			for {
				select {
				case <-ticker.C:
					if onMenu {
						continue
					}

					app.QueueUpdateDraw(func() {
						stops, err := updateStopData(stops, apikey)
						if err != nil {
							panic(err)
						}

						left, right := stopsGetLeftRight(stops, i)
						newframe, err := newTuiStopFrame(stops[i], left, right)
						if err != nil {
							panic(err)
						}

						layout.Clear()
						layout.AddItem(newframe, 0, 1, true)
						app.SetRoot(layout, true)

					})
				}
			}
		}()

		if err := app.SetRoot(layout, true).Run(); err != nil {
			panic(err)
		}
	}
}
