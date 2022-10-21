package main

import (
	"fmt"
	"github.com/koovee/thermia/control"
	"github.com/koovee/thermia/spotprice"
	"os"
	"strconv"
	"time"
)

const (
	defaultThreshold = 5.0
)

type state struct {
	sp        spotprice.State
	cs        control.State
	threshold float64
}

func main() {
	s, err := getEnv()
	if err != nil {
		fmt.Printf("failed to get required environment variables\n")
		return
	}

	err = s.sp.Init()
	if err != nil {
		fmt.Printf("failed to initialize spotprice module")
		return
	}
	err = s.cs.Init()
	if err != nil {
		fmt.Printf("failed to initialize shelly module")
		return
	}

	fmt.Printf("Thermia controller started\n")

	// every even hour
	timer := time.NewTimer(time.Second)

	for {
		select {
		//case <-s.sp.C:
		//	fmt.Printf("spot price update routine failed, exiting..")
		//	return
		case <-timer.C:
			// Update prices
			s.sp.UpdateSpotPrices()

			// Control switch
			price := s.sp.GetPrice(time.Now())
			fmt.Printf("hourly price [%s]: %f\n", time.Now().Format(time.RFC822), price)

			if price <= s.threshold {
				s.cs.SwitchOff()
			} else {
				fmt.Printf("TURN SWITCH: ON (EVU / LOWERED TEMPERATURE) -- NOT REALLY\n")
				s.cs.SwitchOn()
			}

			timer.Reset(time.Now().Truncate(time.Hour).Add(time.Hour).Add(time.Second).Sub(time.Now()))
		}
	}
}

func getEnv() (s state, err error) {
	threshold := os.Getenv("THRESHOLD")
	if threshold != "" {
		s.threshold, err = strconv.ParseFloat(threshold, 64)
		if err != nil {
			fmt.Printf("failed to parse float from environment variable (THRESHOLD): %s\n", err.Error())
			return
		}
	}
	if s.threshold == 0 {
		s.threshold = defaultThreshold
	}

	return
}
