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
	// IsCheapestHour returns true if price is among the n cheapest prices for a given day
	IsCheapestHour(time time.Time, price float64, n int) bool
}

type HourPrices map[string][]float64
