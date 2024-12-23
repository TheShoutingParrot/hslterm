package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/mgutz/ansi"
)

const appName = "hslterm"

const usageText = "hslterm usage:\n\thslterm [OPTIONS]\n\nhslterm lets you see a HSL stop's timetables that update in realtime\n" +
	"Also let's you see a map of the Helsinki metro that updates locations of metros in realtime\n" +
	"\nhslterm requires users to give an api key for https://digitransit.fi/\n" +
	"\t-apikey=APIKEY sets the apikey. Stores it in ~/.config/hslterm/apikey.txt\n" +
	"\t-temp-apikey=APIKEY: sets an api key for the duration of one command\n" +
	"See https://digitransit.fi/en/developers/api-registration/ for instructions on getting a valid api key\n" +
	"\nCommands/options\n" +
	"\t-stop=[NAME OF STOP]: displays the timetable for the next hour, if multiple stops have the same name, will ask user to specify\n" +
	"\t-code=[CODE OF STOP]: specify the code of the stop so no need to specify later (with comma separation you may enter multiple codes)\n" +
	"\t-a: displays/prints all stops and doesn't ask to specify\n" +
	"\t-alerts: prints list of alerts\n" +
	"\t-metro: displays the metro map in terminal (-tui option enabled automatically) [COMING SOON]" +
	"\t-tui: shows the given data in a live updating tui view\n" +
	"\t-api: print current apikey\n" +
	"\t-h/-help: shows this"

func getTerminalWidth() (int, error) {
	var ws struct {
		Rows   uint16
		Cols   uint16
		Xpixel uint16
		Ypixel uint16
	}

	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	if err != 0 {
		return 80, fmt.Errorf("failed to get terminal width: %v", err)
	}

	return int(ws.Cols), nil
}

func hyperlink(text, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

func usageFn(s string) error {
	fmt.Println(usageText)

	os.Exit(0)

	return nil
}

func formatTimeLeft(t int64) string {
	min := int(time.Until(time.Unix(t, 0)).Round(time.Minute).Minutes())
	if min < 1 {
		return "Now"
	} else if min > 120 {
		return fmt.Sprintf("%vh %vmin", min/60)
	} else if min > 60 {
		return fmt.Sprintf("%vh %vmin", min/60, min%60)
	}

	return fmt.Sprintf("%vmin", min)
}

func transportModeEmoji(mode string) string {
	switch mode {
	case "BUS":
		return "🚌"
	case "TRAM":
		return "🚊"
	case "RAIL":
		return "🚆"
	case "SUBWAY":
		return "🚇"
	case "FERRY":
		return "⛴️"
	}

	return "❓"
}

var bold = ansi.ColorFunc("white+b")
var redText = ansi.ColorFunc("red+b")

func main() {
	flag.BoolFunc("help", "usage", usageFn)
	flag.BoolFunc("h", "usage", usageFn)
	flag.BoolFunc("api", "print current apikey", func(s string) error {
		a, err := loadApikey()
		if err != nil {
			fmt.Println("failed: " + err.Error())
			os.Exit(1)
		}
		fmt.Println(a)

		os.Exit(0)

		return nil
	})

	apikey := flag.String("apikey", "", "Sets the API key. Stores it in ~/.config/hslterm/apikey.txt")
	tempApikey := flag.String("temp-apikey", "", "Sets a temporary API key for the duration of one command")
	stop := flag.String("stop", "", "Displays the timetable for the next hour for the given stop")
	code := flag.String("code", "", "Specify the code of the stop to avoid asking later")
	metro := flag.Bool("metro", false, "Displays the metro map in terminal (Enables -tui automatically)")
	alerts := flag.Bool("alerts", false, "prints list of alerts")
	tui := flag.Bool("tui", false, "Shows the given data in a live updating TUI view")
	printAll := flag.Bool("a", false, "Displays/prints all stops and doesn't ask to specify")

	flag.Parse()

	if *tempApikey != "" {
		apikey = tempApikey
	} else if *apikey == "" {
		var err error

		*apikey, err = loadApikey()
		if err != nil {
			fmt.Println(redText("error when loading apikey: " + err.Error()))
			os.Exit(1)
		}
	} else {
		err := saveApiKey(*apikey)
		if err != nil {
			fmt.Println(redText("failed to save api key: " + err.Error()))
			os.Exit(1)
		}
	}

	if *stop != "" {
		// Get stop data
		stops, err := getStopData(*apikey, *stop, 5)
		if err != nil {
			fmt.Println(redText("got err " + err.Error()))
			os.Exit(1)
		}

		if *code != "" {
			*code = strings.ToUpper(*code)
			// separate codes by comma
			if strings.Contains(*code, ",") {
				actualStops := []Stop{}

				codes := strings.Split(*code, ",")
				for _, stop := range stops {
					for _, code := range codes {
						if stop.Code == code {
							actualStops = append(actualStops, stop)
							break
						}
					}
				}

				stops = actualStops
			} else {
				// Find the stop with the given code
				for _, stop := range stops {
					if stop.Code == *code {
						stops = []Stop{stop}
						break
					}
				}
			}
		}

		if *tui {
			tuiDisplayStops(stops, *apikey)

			return
		}

		var selection string
		var reader *bufio.Reader

		// Print all stops if -a is given
		if *printAll {
			goto printAllStops
		}

		// Print the stop with code if code is given
		if *code != "" {
			goto printAllStops
		}

		// Select stops to show
		for i, stop := range stops {
			fmt.Printf("%v) %v (%v, %v) %v\n", i, stop.Name, stop.Desc, stop.Code, transportModeEmoji(stop.VehicleMode))
		}

		fmt.Print("Select stop or press enter for all: ")
		reader = bufio.NewReader(os.Stdin)
		selection, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println(redText("failed to read input: " + err.Error()))
			os.Exit(1)
		}
		selection = strings.TrimSpace(selection)

		// On enter press show all stops
		if selection == "" {
		} else {
			// parse selection and show stop
			i := int([]rune(selection)[0] - '0')
			if i < 0 || i >= len(stops) {
				fmt.Println(redText("invalid selection"))
				os.Exit(1)
			}

			stops = []Stop{stops[i]}
		}

	printAllStops:

		for _, stop := range stops {
			printStop(stop)
			fmt.Print("\n")
		}

		return
	} else if *metro {
		fmt.Println("Coming soon")
		return
	} else if *alerts {
		data, err := getAllAlerts(*apikey)
		if err != nil {
			fmt.Println(redText("got err " + err.Error()))
			os.Exit(1)
		}

		if *tui {
			tuiDisplayAlerts(data)

			return
		}

		printAlerts(data)

		return
	}

	tuiDisplaySearch(*apikey)
}
