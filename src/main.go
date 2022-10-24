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
			now := time.Now()

			// Update prices
			s.sp.UpdateSpotPrices()

			// Control switch
			price := s.sp.GetPrice(now)
			fmt.Printf("hourly price [%s]: %f\n", time.Now().Format(time.RFC822), price)

			if price <= s.threshold {
				// heating ON / NORMAL mode (price is lower than the threshold)
				err = s.cs.SwitchOff()
				if err != nil {
					fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
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

func isCheapestHour(cheapestHours []int) bool {
	hour := time.Now().Hour()
	for _, cheapestHour := range cheapestHours {
		if cheapestHour == hour {
			return true
		}
	}
	return false
}
