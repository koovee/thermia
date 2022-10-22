package control

type Control interface {
	Init(dryRun bool) error
	SwitchOn() error
	SwitchOff() error
}

type HourPrices map[string][]float64
