package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/koovee/thermia/control"
	"github.com/koovee/thermia/spotprice"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimezone = "Europe/Helsinki"
	defaultSchedule = "0,1,2,3,4,5"
)

var version string

type state struct {
	sp          spotprice.State
	cs          control.State
	threshold   float64
	activeHours int
	schedule    map[int]bool
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
				if err != nil {
					err = s.controlBasedOnSchedule()
				}
			} else if s.activeHours > 0 {
				err = s.controlBasedOnActiveHours()
				if err != nil {
					err = s.controlBasedOnSchedule()
				}
			} else if s.threshold > 0 {
				err = s.controlBasedOnThreshold()
				if err != nil {
					err = s.controlBasedOnSchedule()
				}
			} else {
				err = s.controlBasedOnSchedule()
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

	activeHours := os.Getenv("ACTIVE_HOURS")
	if activeHours != "" {
		s.activeHours, err = strconv.Atoi(activeHours)
		if err != nil {
			fmt.Printf("failed to parse int from environment variable (ACTIVE_HOURS): %s\n", err.Error())
			return
		}
	}

	schedule := os.Getenv("SCHEDULE")
	if schedule == "" {
		schedule = defaultSchedule
	}

	s.schedule = make(map[int]bool)
	scheduleString := strings.Split(schedule, ",")
	for _, str := range scheduleString {
		var hour int
		hour, err = strconv.Atoi(str)
		if err != nil {
			fmt.Printf("failed to parse array of integers from environment variable (SCHEDULE): %s\n", err.Error())
			return
		}
		if hour < 0 || hour > 24 {
			err = errors.New("invalid hour")
			fmt.Printf("failed to parse array of integers from environment variable (SCHEDULE): %s\n", err.Error())
			return
		}
		s.schedule[hour] = true
	}
	if len(s.schedule) > 24 {
		err = errors.New("too many hours in schedule")
		fmt.Printf("failed to parse array of integers from environment variable (SCHEDULE): %s\n", err.Error())
		return
	}

	tz := os.Getenv("TZ")
	if tz != "" {
		s.tz = defaultTimezone
	}

	return
}

// controlBasedOnThreshold controls heating based on threshold
func (s state) controlBasedOnThreshold() (err error) {
	now := time.Now()
	price, err := s.sp.GetPrice(now)
	if err != nil {
		fmt.Printf("failed to control based on threshold: %s", err.Error())
		return err
	}

	fmt.Printf("control based on threshold (%.2f)\n", s.threshold)
	fmt.Printf("hourly price [%s]: %.2f\n", time.Now().Format(time.RFC822), price)

	if price <= s.threshold {
		// heating ON / NORMAL mode (price is lower than the threshold)
		fmt.Printf("Heating ON: price lower than the threshold: %0.2f (threshold: %0.2f)\n", price, s.threshold)

		err = s.cs.SwitchOff()
		if err != nil {
			fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
			return err
		}
	} else {
		// heating OFF / ROOM LOWERING mode
		fmt.Printf("Heating OFF: price higher than the threshold: %0.2f (threshold: %0.2f)\n", price, s.threshold)
		err = s.cs.SwitchOn()
		if err != nil {
			fmt.Printf("failed to turn heat pump off / room lowering mode: %s\n", err.Error())
		}
	}
	return nil
}

// controlBasedOnActiveHours controls heating based on activeHours
func (s state) controlBasedOnActiveHours() (err error) {
	now := time.Now()
	price, err := s.sp.GetPrice(now)
	if err != nil {
		fmt.Printf("failed to control based on activeHours: %s", err.Error())
		return err
	}

	if isCheapestHour(s.sp.CheapestHours(s.activeHours)) {
		// Heating ON / NORMAL mode (this is one of the cheapest hours)
		fmt.Printf("Heating ON: this is one of the %d cheapest hours: %0.2f\n", s.activeHours, price)
		err = s.cs.SwitchOff()
		if err != nil {
			fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
			return err
		}
	} else {
		fmt.Printf("Heating OFF: this is not one of the %d cheapest hours\n", s.activeHours)
		err = s.cs.SwitchOn()
		if err != nil {
			fmt.Printf("failed to turn heat pump off / room lowering mode: %s\n", err.Error())
		}
	}

	return nil
}

// controlBasedOnThresholdAndActiveHours controls heating based on threshold and activeHours
func (s state) controlBasedOnThresholdAndActiveHours() (err error) {
	now := time.Now()
	price, err := s.sp.GetPrice(now)
	if err != nil {
		fmt.Printf("failed to control based on threshold and activehours: %s", err.Error())
		return err
	}

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
				// heating OFF / ROOM LOWERING mode
				err = s.cs.SwitchOn()
				if err != nil {
					fmt.Printf("failed to turn heat pump off / room lowering mode: %s\n", err.Error())
				}
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

// controlBasedOnCron controls heating based on cron
func (s state) controlBasedOnSchedule() (err error) {
	now := time.Now()
	price, _ := s.sp.GetPrice(now)

	fmt.Printf("control based on schedule\n")

	if s.schedule[now.Hour()] == true {
		// Heating ON / NORMAL mode
		fmt.Printf("Heating ON: schedule (price: %0.2f)\n", price)
		err = s.cs.SwitchOff()
		if err != nil {
			fmt.Printf("failed to turn heat pump on: %s\n", err.Error())
			return err
		}
	} else {
		// heating OFF / ROOM LOWERING mode
		fmt.Printf("Heating OFF: schedule (price: %0.2f)\n", price)
		err = s.cs.SwitchOn()
		if err != nil {
			fmt.Printf("failed to turn heat pump off / room lowering mode: %s\n", err.Error())
		}

	}
	return nil
}
