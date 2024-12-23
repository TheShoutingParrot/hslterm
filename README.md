# hslterm

Unofficial CLI application that displays HSL stop info and alerts.

## Installation

If your go is setup properly the following should work after downloading/cloning this repo:

`go install .`

After that you should be able to run `hslterm`

## Usage

You must have a valid digitransit api key to use this software.

See https://digitransit.fi/en/developers/api-registration/ for instructions on getting a valid api key.

Run:

`hslterm -apikey=APIKEY`

to set your apikey.

To get the timetable of a stop run:

`hslterm -stop=[NAME OF STOP]`

This will first ask which stop you want to see. Run it with the -a flag to get every stop's data.

To see realtime updating data in a tui view you can run:

`hslterm -stop=[NAME OF STOP] -tui`

You can specify a hsl stop code like so:

`hslterm -stop=[NAME OF STOP] -code=[CODE OF STOP]`

Example output when running with `hslterm -stop=Aalto -code=E0003`:
```
Stop: Aalto-yliopisto (M) (Otaniementie 12, E0003) 🚇
Location: 60.184516, 24.823515
Routes: 
🚇      M2 - Tapiola - Mellunmäki
🚇      M1 - Kivenlahti - Vuosaari

╭───────────────────────────────────┬───────────┬───────────╮
│ ROUTE                             │ DEPARTING │ TIME LEFT │
├───────────────────────────────────┼───────────┼───────────┤
│ M2 - Mellunmäki via Rautatientori │ 12:25     │ 2min      │
│ M1 - Vuosaari via Rautatientori   │ 12:28     │ 5min      │
│ M2 - Mellunmäki via Rautatientori │ 12:32     │ 9min      │
│ M1 - Vuosaari via Rautatientori   │ 12:35     │ 12min     │
│ M2 - Mellunmäki via Rautatientori │ 12:40     │ 17min     │
╰───────────────────────────────────┴───────────┴───────────╯
```

You can also view ongoing alerts/infos by running:

`hslterm -alerts`

And for a nicer view run with `-tui`.



### Rofi script in scripts/

I've made a neat script for myself. I've included it in [scripts/rofi_stop_selector.sh](scripts/rofi_stop_selector.sh).

See a gif on what it does:

![rofi script in action](assets/rofi_script_usage.gif)
