package entsoe

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	apiUrl     = "https://transparency.entsoe.eu/api"
	DateLayout = "20060102"
)

type State struct {
	token     string
	threshold float64
	prices    map[string][]float64
}

type A44Response struct {
	XMLName    xml.Name `xml:"Publication_MarketDocument"`
	TimeSeries []struct {
		XMLName xml.Name `xml:"TimeSeries"`
		MRID    string   `xml:"mRID"`
		Period  struct {
			XMLName      xml.Name `xml:"Period"`
			TimeInterval struct {
				XMLName xml.Name `xml:"timeInterval"`
				Start   string   `xml:"start"`
				End     string   `xml:"end"`
			}
			Resolution string `xml:"resolution"`
			Point      []struct {
				XMLName  xml.Name `xml:"Point"`
				Position int      `xml:"position"`
				Price    string   `xml:"price.amount"`
			}
		}
	}
}

var s State

func Init() State {
	err := s.getEnv()
	if err != nil {
		fmt.Printf("failed to get required environment variables")
		return State{}
	}
	return s
}

func (s *State) getEnv() error {
	s.token = os.Getenv("TOKEN")
	if s.token == "" {
		return errors.New("TOKEN not set")
	}
	return nil
}

// updateSpotPrices ..
func (s *State) UpdateSpotPrices(ch chan bool) {
	var retryCount = 0
	httpClient := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	defer close(ch)

	s.prices = make(map[string][]float64)
	loc, _ := time.LoadLocation("Europe/Helsinki")
	timer := time.NewTimer(time.Second)
	<-timer.C

	for {
		fmt.Printf("getting spot prices from %s\n", apiUrl)

		// delete yesterdays records
		fmt.Printf("DEBUG: map size before cleanup: %d\n", len(s.prices))
		delete(s.prices, time.Now().Add(-24*time.Hour).Format(DateLayout))
		fmt.Printf("DEBUG: map size after cleamup: %d\n", len(s.prices))

		req, err := http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			fmt.Printf("Failed to create http request")
			retryCount++
			time.Sleep(time.Second * time.Duration(retryCount*retryCount))
			continue
		}
		retryCount = 0

		today := time.Now().Format("20060102")
		periodStart := today + "0000"
		periodEnd := today + "0100"

		q := url.Values{}
		q.Add("securityToken", s.token)
		q.Add("documentType", "A44")
		q.Add("In_domain", "10YFI-1--------U")
		q.Add("out_domain", "10YFI-1--------U")
		q.Add("periodStart", periodStart)
		q.Add("periodEnd", periodEnd)
		req.URL.RawQuery = q.Encode()

		resp, err := httpClient.Do(req)

		fmt.Printf("DEBUG: %s\n", req.URL)
		fmt.Printf("DEBUG: %v\n", resp.Status)

		foo := A44Response{}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read http response body")
			return
		}
		//fmt.Printf("body: %s\n", body)

		if err = xml.Unmarshal(body, &foo); err != nil {
			fmt.Printf("failed to unmarshal xml")
			return
		}

		for _, v := range foo.TimeSeries {
			var prices [24]float64
			for _, v := range v.Period.Point {
				prices[v.Position-1], err = strconv.ParseFloat(v.Price, 64)
				if err != nil {
					fmt.Printf("failed to convert price to float")
					return
				}
			}
			s.prices[time.Now().Format(DateLayout)] = prices[:]
		}

		fmt.Printf("DEBUG: %v\n", s.prices)

		resp.Body.Close()

		now := time.Now()
		fmt.Printf("DEBUG: Sleeping %s\n", time.Date(now.Year(), now.Month(), now.Add(24*time.Hour).Day(), 0, 0, 1, 0, loc).Sub(time.Now()))
		timer.Reset(time.Date(now.Year(), now.Month(), now.Add(24*time.Hour).Day(), 0, 0, 1, 0, loc).Sub(time.Now()))
		<-timer.C
	}
}
