package spotprice

import (
	"time"
)

type SpotPrice interface {
	Init() error
	// GetPrice returns price for a given time
	GetPrice(time time.Time) float64
	// UpdatePrices retrieves price updates from 3rd party provider
	UpdatePrices() error
	// CheapestHours returns the cheapest n hours for a given day
	CheapestHours(n int) []int
}

type HourPrices map[string][]float64

func IsCheapestHour(hour int, cheapestHours []int) bool {
	for _, cheapestHour := range cheapestHours {
		if cheapestHour == hour {
			return true
		}
	}
	return false
}
