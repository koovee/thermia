package spotprice

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	apiUrl     = "https://transparency.entsoe.eu/api"
	DateLayout = "20060102"
	highPrice  = 9999.99
)

type State struct {
	token     string
	threshold float64
	HourPrice HourPrices
	C         chan bool
	M         *sync.Mutex
	hc        http.Client
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

func (s *State) Init() (err error) {
	err = s.getEnv()
	if err != nil {
		fmt.Printf("failed to get required environment variables: %s\n", err.Error())
		return err
	}

	s.hc = http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	s.C = make(chan bool)
	s.M = &sync.Mutex{}
	s.HourPrice = make(map[string][]float64)

	return nil
}

// GetPrice returns price in c/kWh
func (s State) GetPrice(time time.Time) (float64, error) {
	s.M.Lock()
	defer s.M.Unlock()
	hour := time.Hour()
	if len(s.HourPrice[time.Format(DateLayout)]) == 0 {
		fmt.Printf("no pricing available for %s hour %d\n", time.String(), hour)
		return 0, errors.New("no price information available")
	}
	return s.HourPrice[time.Format(DateLayout)][hour] / 10, nil
}

// UpdateSpotPrices ..
func (s *State) UpdateSpotPrices() {
	var retryCount = 0

	yesterday := time.Now().Add(-24 * time.Hour).Format(DateLayout)
	day := time.Now().Format(DateLayout)
	tomorrow := time.Now().Add(24 * time.Hour).Format(DateLayout)

	s.M.Lock()
	defer s.M.Unlock()

	periodStart := day + "0000"
	periodEnd := day + "0100"

	if time.Now().Hour() > 18 && len(s.HourPrice[day]) > 0 {
		if len(s.HourPrice[tomorrow]) == 0 {
			periodStart = tomorrow + "0000"
			periodEnd = tomorrow + "0100"
			day = tomorrow
			// TODO: change variable name...
		} else {
			// enough pricing data in store..
			return
		}
	} else if len(s.HourPrice[day]) > 0 {
		// enough pricing data in store..
		return
	}

	fmt.Printf("getting spot prices from %s\n", apiUrl)

	// delete yesterdays records
	fmt.Printf("DEBUG: map size before cleanup: %d\n", len(s.HourPrice))
	delete(s.HourPrice, yesterday)
	fmt.Printf("DEBUG: map size after cleanup: %d\n", len(s.HourPrice))

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		fmt.Printf("Failed to create http request\n")
		retryCount++
		time.Sleep(time.Second * time.Duration(retryCount*retryCount))
		return
	}
	retryCount = 0

	q := url.Values{}
	q.Add("securityToken", s.token)
	q.Add("documentType", "A44")
	q.Add("In_domain", "10YFI-1--------U")
	q.Add("out_domain", "10YFI-1--------U")
	q.Add("periodStart", periodStart)
	q.Add("periodEnd", periodEnd)
	req.URL.RawQuery = q.Encode()

	resp, err := s.hc.Do(req)
	if err != nil {
		fmt.Printf("failed to make http request: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: %s\n", req.URL)
	fmt.Printf("DEBUG: %v\n", resp.Status)

	hourlyPrices := A44Response{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read http response body\n")
		return
	}
	//fmt.Printf("body: %s\n", body)

	if err = xml.Unmarshal(body, &hourlyPrices); err != nil {
		fmt.Printf("failed to unmarshal xml\n")
		return
	}

	for _, v := range hourlyPrices.TimeSeries {
		var p []float64

		// add dummy data for offset from utc hours
		_, offset := time.Now().Zone()
		offsetHours := offset / 3600
		for i := 0; i < offset/3600; i++ {
			fmt.Printf("foo\n")
			p = append(p, highPrice)
		}

		for _, v := range v.Period.Point {
			if v.Position > 24 {
				fmt.Printf("DEBUG: v.Position larger than 24 -- check entsoe response!!\n")
				fmt.Printf("DEBUG response body: %s\n", body)
			} else {
				price, err := strconv.ParseFloat(v.Price, 64)
				if err != nil {
					fmt.Printf("failed to convert price to float\n")
					return
				}
				p = append(p, price)
			}
		}
		s.HourPrice[day] = p[offsetHours-1 : 24+offsetHours-1]
		s.HourPrice[tomorrow] = p[24+offsetHours-1:]
	}

	fmt.Printf("DEBUG: %v\n", s.HourPrice)
}

// CheapestHours returns the cheapest n hours for a given day
func (s State) CheapestHours(n int) (cheapestPrices []int) {
	var cheapestIndex int
	var cheapest float64

	prices := s.HourPrice[time.Now().Format(DateLayout)]

	for i := 0; i < n; i++ {
		cheapest = highPrice
		for i, price := range prices {
			if price < cheapest {
				if len(cheapestPrices) > 0 {
					if price <= prices[cheapestPrices[len(cheapestPrices)-1]] {
						continue
					}
				}
				cheapest = price
				cheapestIndex = i
			}
		}
		cheapestPrices = append(cheapestPrices, cheapestIndex)
	}
	return cheapestPrices
}

func (s *State) getEnv() error {
	s.token = os.Getenv("TOKEN")
	if s.token == "" {
		return errors.New("TOKEN not set")
	}
	return nil
}
