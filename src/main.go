package main

import (
	"flag"
	"fmt"
	"github.com/koovee/thermia/control"
	"github.com/koovee/thermia/spotprice"
	"os"
	"strconv"
	"time"
)

const (
	defaultTimezone    = "Europe/Helsinki"
	defaultThreshold   = 10.0
	defaultActiveHours = 6
)

var version string

type state struct {
	sp          spotprice.State
	cs          control.State
	threshold   float64
	activeHours int
	tz          string
}

func main() {

	dryRun := flag.Bool("dryrun", false, "disable relay control")
	flag.Parse()

	s, err := getEnv()
	if err != nil {
		fmt.Printf("failed to get required environment variables\n")
		return
	}

	_, err = time.LoadLocation(s.tz)
	if err != nil {
		fmt.Printf("failed to set timezone (%s): %s\n", s.tz, err.Error())
		return
	}

	err = s.sp.Init()
	if err != nil {
		fmt.Printf("failed to initialize spotprice module\n")
		return
	}
	err = s.cs.Init(*dryRun)
	if err != nil {
		fmt.Printf("failed to initialize shelly module\n")
		return
	}

	fmt.Printf("Thermia controller started (version: %s, dryRun: %v, treshold: %0.2f, activeHours: %d)\n", version, *dryRun, s.threshold, s.activeHours)

	// every even hour
	timer := time.NewTimer(time.Second)

	for {
		select {
		case <-timer.C:
			// Update prices
			s.sp.UpdateSpotPrices()

			// Control relay based on configuration and hourly price
			if s.activeHours > 0 && s.threshold > 0 {
				err = s.controlBasedOnThresholdAndActiveHours()
			} else if s.activeHours > 0 {
				err = s.controlBasedOnActiveHours()
			} else if s.threshold > 0 {
				err = s.controlBasedOnThreshold()
			} else {
				err = s.controlBasedOnCron()
			}
			if err != nil {
				fmt.Printf("failed to control relay: %s\n", err.Error())
			}

			timer.Reset(time.Now().Truncate(time.Hour).Add(time.Hour).Add(time.Second).Sub(time.Now()))
		}
	}
}

func isCheapestHour(cheapestHours []int) bool {
	hour := time.Now().Hour()
	for _, cheapestHour := range cheapestHours {
		if cheapestHour == hour {
			return true
		}
	}
	return false
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

	activeHours := os.Getenv("ACTIVE_HOURS")
	if activeHours != "" {
		s.activeHours, err = strconv.Atoi(activeHours)
		if err != nil {
			fmt.Printf("failed to parse int from environment variable (ACTIVE_HOURS): %s\n", err.Error())
			return
		}
	} else {
		s.activeHours = defaultActiveHours
	}

	tz := os.Getenv("TZ")
	if tz != "" {
		s.tz = defaultTimezone
	}

	return
}

func (s state) controlBasedOnThreshold() (err error) {
	return nil
}

func (s state) controlBasedOnActiveHours() (err error) {
	return nil
}

func (s state) controlBasedOnThresholdAndActiveHours() (err error) {
	now := time.Now()
	price := s.sp.GetPrice(now)

	fmt.Printf("control based on threshold (%.2f) and active hours (%d)\n", s.threshold, s.activeHours)
	fmt.Printf("hourly price [%s]: %.2f\n", time.Now().Format(time.RFC822), price)

	if price <= s.threshold {
		// heating ON / NORMAL mode (price is lower than the threshold)
		err = s.cs.SwitchOff()
		if err != nil {
			fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
			return err
		}
	} else {
		// price is higher than the threshold
		if s.activeHours > 0 {
			if isCheapestHour(s.sp.CheapestHours(s.activeHours)) {
				// Heating ON / NORMAL mode (this is one of the cheapest hours)
				fmt.Printf("Heating ON: price higher than threshold but this is one of the %d cheapest hours: %0.2f\n", s.activeHours, price)
				err = s.cs.SwitchOff()
				if err != nil {
					fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
					return err
				}
			} else {
				fmt.Printf("Heating OFF: price higher than threshold and this is not one of the %d cheapest hours\n", s.activeHours)
			}
		} else {
			// heating OFF / ROOM LOWERING mode
			err = s.cs.SwitchOn()
			if err != nil {
				fmt.Printf("failed to turn heat pump off / room lowering mode: %s\n", err.Error())
			}
		}
	}
	return nil
}

func (s state) controlBasedOnCron() (err error) {
	return nil
}
