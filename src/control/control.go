package control

type Control interface {
	Init() error
	SwitchOn() error
	SwitchOff() error
}

type HourPrices map[string][]float64
