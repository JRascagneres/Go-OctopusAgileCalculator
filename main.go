package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"time"
)

func main() {
	calcRuntimeCosts(RuntimeCostInput{
		InputItems: []RuntimeCostInputItem{
			{
				Name:  "Battery charge",
				Hours: 1.5,
			},
		},
	})

	return
}

type RuntimeCostInput struct {
	InputItems []RuntimeCostInputItem
}

type RuntimeCostInputItem struct {
	Name  string
	Hours float64
}

func calcRuntimeCosts(runtimeInputs RuntimeCostInput) {
	ratesResponse, err := getRates()
	if err != nil {
		fmt.Println(err)
		return
	}

	future := filterToFutureOnly(ratesResponse.Results)
	sortByDate(future)

	for _, runtimeInput := range runtimeInputs.InputItems {
		lowestTime, lowestEndTime, lowestCost := calcRuntimeCost(future, runtimeInput.Hours)
		fmt.Printf("%s: %s - %s: %.2fp\n", runtimeInput.Name, timeInFormat(lowestTime), timeInFormat(lowestEndTime), lowestCost)
	}
}

func timeInFormat(t time.Time) string {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(err)
	}
	return t.In(loc).Format("15:04")
}

func calcRuntimeCost(futureRates []Rate, runtimeHours float64) (time.Time, time.Time, float64) {
	runtimeSegments := int(math.Ceil(runtimeHours / 0.5))

	lowestTime := futureRates[0].ValidFrom
	lowestCost := float64(1000)
	startCost := float64(1000)

	for i := range futureRates {
		if i+runtimeSegments > len(futureRates) {
			break
		}
		currentCostRate := float64(0)
		for j := 0; j < runtimeSegments; j++ {
			currentCostRate += futureRates[i+j].ValueIncVat / 2
		}
		if currentCostRate < lowestCost {
			startCost = futureRates[i].ValueIncVat
			lowestCost = currentCostRate
			lowestTime = futureRates[i].ValidFrom
		}
	}

	lowestEndTime := lowestTime.Add(time.Duration(runtimeSegments) * time.Minute * 30)

	fmt.Println(startCost)
	return lowestTime, lowestEndTime, lowestCost
}

type RatesResponse struct {
	Count   int    `json:"count"`
	Results []Rate `json:"results"`
}

type Rate struct {
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	ValueIncVat float64   `json:"value_inc_vat"`
}

func sortByDate(rates []Rate) {
	sort.Slice(rates, func(i, j int) bool {
		return rates[i].ValidFrom.Before(rates[j].ValidFrom)
	})
}

func filterToFutureOnly(rates []Rate) []Rate {
	var filteredRates []Rate
	now := time.Now()

	for _, rate := range rates {
		if rate.ValidFrom.After(now) {
			filteredRates = append(filteredRates, rate)
		}
	}

	return filteredRates
}

func getRates() (*RatesResponse, error) {
	url := "https://api.octopus.energy/v1/products/AGILE-FLEX-22-11-25/electricity-tariffs/E-1R-AGILE-FLEX-22-11-25-D/standard-unit-rates/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var ratesResponse RatesResponse

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&ratesResponse)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &ratesResponse, nil
}
