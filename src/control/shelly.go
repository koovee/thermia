package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	defaultShellyUrl = "http://10.0.0.84/relay/0"
)

type State struct {
	url    string
	hc     *http.Client
	dryRun bool
}

type statusResponse struct {
	Ison           bool    `json:"ison"`
	HasTimer       bool    `json:"has_timer"`
	TimerStartedAt int     `json:"timer_started_at"`
	TimerDuration  float64 `json:"timer_duration"`
	TimerRemaining float64 `json:"timer_remaining"`
	Source         string  `json:"source"`
}

var s State

func (s *State) Init(dryRun bool) error {
	err := s.getEnv()
	if err != nil {
		fmt.Printf("failed to get required environment variables")
		return err
	}
	s.dryRun = dryRun
	return nil
}

// SwitchOff turns switch OFF which means Thermia is operating in NORMAL mode
func (s State) SwitchOff() error {
	var response statusResponse

	// check current state
	resp, err := http.Get(s.url)
	if err != nil {
		fmt.Printf("Failed to create http request")
		return errors.New("failed to create http request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)                  // response body is []byte
	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Printf("Can not unmarshal JSON: %s\n", err.Error())
		return errors.New("failed to unmarshal JSON response")
	}
	if response.Ison == true {
		// change state
		if s.dryRun {
			fmt.Printf("DRY RUN -- Switch is on, turning it off (NORMAL OPERATION) -- DRY RUN\n")
		} else {
			fmt.Printf("Switch is on, turning it off (NORMAL OPERATION)\n")
			resp, err := http.Get(s.url + "?turn=off")
			if err != nil {
				fmt.Printf("Failed to create http request")
				return errors.New("failed to create http request")
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("failed to set switch off: %s\n", err.Error())
				return errors.New("failed to set switch off")
			}
		}
	}

	return nil
}

// SwitchOn tunrs switch ON which means Thermia is operating in heat reduction mode (normal-2 degress)
func (s State) SwitchOn() error {
	var response statusResponse

	// check current state
	resp, err := http.Get(s.url)
	if err != nil {
		fmt.Printf("Failed to create http request")
		return errors.New("failed to create http request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)                  // response body is []byte
	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Printf("Can not unmarshal JSON: %s\n", err.Error())
		return errors.New("failed to unmarshal JSON response")
	}
	if response.Ison == false {
		// change state
		if s.dryRun {
			fmt.Printf("DRY RUN -- Switch is off, turning it on (EVU ON / LOWERED TEMPERATURE) -- DRY RUN\n")
		} else {
			fmt.Printf("Switch is off, turning it on (EVU ON / LOWERED TEMPERATURE)\n")
			resp, err := http.Get(s.url + "?turn=on")
			if err != nil {
				fmt.Printf("Failed to create http request")
				return errors.New("failed to create http request")
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("failed to set switch on: %s\n", err.Error())
				return errors.New("failed to set switch on")
			}
		}
	}

	return nil
}

func (s *State) getEnv() error {
	s.url = os.Getenv("SHELLY_URL")
	if s.url == "" {
		s.url = defaultShellyUrl
	}
	return nil
}
