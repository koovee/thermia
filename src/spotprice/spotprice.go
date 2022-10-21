package spotprice

import (
	"time"
)

type SpotPrice interface {
	Init() error
	GetPrice(time time.Time) float64
	UpdatePrices() error
}

type HourPrices map[string][]float64
