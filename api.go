package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ApiRequest(apikey string, jsonquery string, data any) error {
	reqBody := strings.NewReader(jsonquery)

	httpReq, err := http.NewRequest("POST", "https://api.digitransit.fi/routing/v1/routers/hsl/index/graphql", reqBody)
	if err != nil {
		return err
	}

	httpReq.Header.Set("digitransit-subscription-key", apikey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", httpResp.StatusCode)
	}

	err = json.NewDecoder(httpResp.Body).Decode(&data)
	if err != nil {
		return err
	}

	return nil
}

type Alert struct {
	AlertCause         string `json:"alertCause"`
	AlertEffect        string `json:"alertEffect"`
	AlertHash          int    `json:"alertHash"`
	AlertHeaderText    string `json:"alertHeaderText"`
	AlertSeverityLevel string `json:"alertSeverityLevel"`
	AlertUrl           string `json:"alertUrl"`
	EffectiveStartDate int64  `json:"effectiveStartDate"`
	EffectiveEndDate   int64  `json:"effectiveEndDate"`
	Feed               string `json:"feed"`
	ID                 string `json:"id"`
}

func getAllAlerts(apikey string) ([]Alert, error) {
	var data struct {
		Data struct {
			Alerts []Alert `json:"alerts"`
		} `json:"data"`
	}

	query := `{"query": "query { alerts { alertCause alertEffect alertHeaderText alertSeverityLevel alertUrl effectiveStartDate effectiveEndDate feed id } }"}`

	err := ApiRequest(apikey, query, &data)

	return data.Data.Alerts, err
}

type Route struct {
	LongName  string `json:"longName"`
	ShortName string `json:"shortName"`
	Mode      string `json:"mode"`
	Url       string `json:"url"`
}

type StopTimes struct {
	Headsign           string `json:"headsign"`
	RealtimeState      string `json:"realtimeState"`
	ScheduledArrival   int64  `json:"scheduledArrival"`
	RealtimeArrival    int64  `json:"realtimeArrival"`
	ScheduledDeparture int64  `json:"scheduledDeparture"`
	RealtimeDeparture  int64  `json:"realtimeDeparture"`
	ServiceDay         int64  `json:"serviceDay"`
	Trip               struct {
		RouteShortName string `json:"routeShortName"`
	} `json:"trip"`
}

type Stop struct {
	Alerts      []Alert     `json:"alerts"`
	Code        string      `json:"code"`
	Desc        string      `json:"desc"`
	Direction   string      `json:"direction"`
	Lat         float64     `json:"lat"`
	Lon         float64     `json:"lon"`
	Name        string      `json:"name"`
	Routes      []Route     `json:"routes"`
	StopTimes   []StopTimes `json:"stoptimesWithoutPatterns"`
	VehicleMode string      `json:"vehicleMode"`
	GtfsID      string      `json:"gtfsId"`
}

func getStopData(apikey string, stopName string, stopTimesN int) ([]Stop, error) {
	var data struct {
		Data struct {
			Stops []Stop `json:"stops"`
		} `json:"data"`
	}

	query := fmt.Sprintf(`{"query": "query { stops(name: \"%v\") { alerts { alertCause alertEffect alertHeaderText alertSeverityLevel alertUrl effectiveStartDate effectiveEndDate feed id } code desc direction lat lon name vehicleMode gtfsId routes { longName shortName mode url } stoptimesWithoutPatterns (numberOfDepartures: %v) { headsign realtimeState scheduledArrival realtimeArrival scheduledDeparture realtimeDeparture serviceDay trip { routeShortName } } } }"}`, stopName, stopTimesN)

	err := ApiRequest(apikey, query, &data)

	return data.Data.Stops, err
}

func (s *Stop) refresh(apikey string) error {
	var data struct {
		Data struct {
			Stop Stop `json:"stop"`
		} `json:"data"`
	}

	query := fmt.Sprintf(`{"query": "query { stop(id: \"%v\") { alerts { alertCause alertEffect alertHeaderText alertSeverityLevel alertUrl effectiveStartDate effectiveEndDate feed id } code desc direction lat lon name vehicleMode gtfsId routes { longName shortName mode url } stoptimesWithoutPatterns { headsign realtimeState scheduledArrival realtimeArrival scheduledDeparture realtimeDeparture serviceDay trip { routeShortName } } } }"}`, s.GtfsID)

	err := ApiRequest(apikey, query, &data)
	if err != nil {
		return err
	}

	*s = data.Data.Stop

	return nil
}

func updateStopData(s []Stop, apikey string) ([]Stop, error) {
	for i := range s {
		err := s[i].refresh(apikey)
		if err != nil {
			return s, err
		}
	}

	return s, nil
}
