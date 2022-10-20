package main

import (
	"fmt"
	"github.com/koovee/thermia/entsoe"
	"github.com/koovee/thermia/shelly"
	"os"
	"strconv"
	"time"
)

const (
	defaultThreshold = 5.0
)

type state struct {
	entsoe    entsoe.State
	shelly    shelly.State
	threshold float64
	prices    map[string][]float64
}

var s state

func main() {
	s.getEnv()
	s.entsoe = entsoe.Init()
	s.shelly = shelly.Init()

	// update spot prices once a day
	ch := make(chan bool)
	go s.entsoe.UpdateSpotPrices(ch)

	// every even hour
	timer := time.NewTimer(time.Now().Truncate(time.Hour).Add(time.Hour).Add(time.Second).Sub(time.Now()))
	for {
		select {
		case <-ch:
			fmt.Printf("spot price update routine failed, exiting..")
			return
		case <-timer.C:
			price := s.getHourPrice(time.Now())
			fmt.Printf("hourly price [%s]: %f\n", time.Now().Format(time.RFC822), price)

			if price <= s.threshold {
				shelly.SwitchOff()
			} else {
				fmt.Printf("TURN SWITCH: ON (EVU / LOWERED TEMPERATURE) -- NOT REALLY\n")
				shelly.SwitchOn()
			}

			timer.Reset(time.Now().Truncate(time.Hour).Add(time.Hour).Add(time.Second).Sub(time.Now()))
		}
	}
}

func (s *state) getEnv() {
	threshold := os.Getenv("THRESHOLD")
	if threshold != "" {
		s.threshold, _ = strconv.ParseFloat(threshold, 64)
	}
	if s.threshold == 0 {
		s.threshold = defaultThreshold
	}
}

// getHourPrice returns price in c/kWh
func (s state) getHourPrice(time time.Time) float64 {
	hour := time.Hour()
	if len(s.prices[time.Format(entsoe.DateLayout)]) < hour-1 {
		fmt.Printf("no pricing available for %s hour %d\n", time.String(), hour)
		return 0
	}
	return s.prices[time.Format(entsoe.DateLayout)][hour] / 10
}
