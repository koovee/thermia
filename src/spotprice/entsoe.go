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
		fmt.Printf("failed to get required environment variables")
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
func (s State) GetPrice(time time.Time) float64 {
	s.M.Lock()
	defer s.M.Unlock()
	hour := time.Hour()
	if len(s.HourPrice[time.Format(DateLayout)]) < hour-1 {
		fmt.Printf("no pricing available for %s hour %d\n", time.String(), hour)
		return 0
	}
	return s.HourPrice[time.Format(DateLayout)][hour] / 10
}

// UpdateSpotPrices ..
func (s *State) UpdateSpotPrices() {
	var retryCount = 0

	yesterday := time.Now().Add(-24 * time.Hour).Format(DateLayout)
	today := time.Now().Format(DateLayout)
	tomorrow := time.Now().Add(24 * time.Hour).Format(DateLayout)

	s.M.Lock()
	defer s.M.Unlock()

	periodStart := today + "0000"
	periodEnd := today + "0100"

	if time.Now().Hour() > 18 {
		if len(s.HourPrice[tomorrow]) == 0 {
			periodStart = tomorrow + "0000"
			periodEnd = tomorrow + "0100"
		} else {
			// enough pricing data in store..
			return
		}
	}

	fmt.Printf("getting spot prices from %s\n", apiUrl)

	// delete yesterdays records
	fmt.Printf("DEBUG: map size before cleanup: %d\n", len(s.HourPrice))
	delete(s.HourPrice, yesterday)
	fmt.Printf("DEBUG: map size after cleamup: %d\n", len(s.HourPrice))

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		fmt.Printf("Failed to create http request")
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
	defer resp.Body.Close()

	fmt.Printf("DEBUG: %s\n", req.URL)
	fmt.Printf("DEBUG: %v\n", resp.Status)

	hourlyPrices := A44Response{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read http response body")
		return
	}
	//fmt.Printf("body: %s\n", body)

	if err = xml.Unmarshal(body, &hourlyPrices); err != nil {
		fmt.Printf("failed to unmarshal xml")
		return
	}

	for _, v := range hourlyPrices.TimeSeries {
		var p [24]float64
		for _, v := range v.Period.Point {
			p[v.Position-1], err = strconv.ParseFloat(v.Price, 64)
			if err != nil {
				fmt.Printf("failed to convert price to float")
				return
			}
		}
		s.HourPrice[today] = p[:]
	}

	fmt.Printf("DEBUG: %v\n", s.HourPrice)
}

func (s *State) getEnv() error {
	s.token = os.Getenv("TOKEN")
	if s.token == "" {
		return errors.New("TOKEN not set")
	}
	return nil
}
