package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

func main() {
	calcRuntimeCosts([]RuntimeCostInput{
		{
			Name:  "3D Print Squares",
			Hours: 9,
		},
		{
			Name:  "Washer + Dryer",
			Hours: 4,
		},
	})

	return
}

type RuntimeCostInput struct {
	Name  string
	Hours int
}

func calcRuntimeCosts(runtimeInputs []RuntimeCostInput) {
	ratesResponse, err := getRates()
	if err != nil {
		fmt.Println(err)
		return
	}

	future := filterToFutureOnly(ratesResponse.Results)
	sortByDate(future)

	for _, runtimeInput := range runtimeInputs {
		lowestTime, lowestCost := calcRuntimeCost(future, runtimeInput.Hours)
		fmt.Printf("Lowest cost for %s is %f at %s\n", runtimeInput.Name, lowestCost, lowestTime)
	}
}

func calcRuntimeCost(futureRates []Rate, runtimeHours int) (time.Time, float64) {
	runtimeSegments := runtimeHours * 2

	lowestTime := futureRates[0].ValidFrom
	lowestCost := float64(1000)

	for i := range futureRates {
		if i+runtimeSegments > len(futureRates) {
			break
		}
		currentCostRate := float64(0)
		for j := 0; j < runtimeSegments; j++ {
			currentCostRate += futureRates[i+j].ValueIncVat
		}
		if currentCostRate < lowestCost {
			lowestCost = currentCostRate
			lowestTime = futureRates[i].ValidFrom
		}
	}

	return lowestTime, lowestCost
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

	for _, rate := range rates {
		if rate.ValidFrom.After(time.Now()) {
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
