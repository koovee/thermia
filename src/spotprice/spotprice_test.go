package spotprice

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	s := State{}

	if err := s.Init(); err == nil {
		t.Errorf("init() should have failed, but it succeeded")
	}

	os.Setenv("TOKEN", "12345")
	if err := s.Init(); err != nil {
		t.Errorf("init() with TOKEN set did not succeed")
	}
	os.Unsetenv("TOKEN")
}

func TestCheapestHours(t *testing.T) {
	s := State{}
	s.HourPrice = make(map[string][]float64)
	set1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0, 17.0, 18.0, 19.0, 20.0, 21.0, 22.0, 23.0, 24.0}
	set2 := []float64{-5.0, -4.0, -3.0, -2.0, -1.0, 0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0}

	cases := map[string]struct {
		hourPrice      []float64
		hour           int
		hours          int
		expectedResult bool
	}{
		"Set1: Is hour 0 the cheapest hour":            {hourPrice: set1, hour: 0, hours: 1, expectedResult: true},
		"Set1: Is hour 1 the cheapest hour":            {hourPrice: set1, hour: 1, hours: 1, expectedResult: false},
		"Set1: Is hour 4 on of the 5 cheapest hours":   {hourPrice: set1, hour: 4, hours: 5, expectedResult: true},
		"Set1: Is hour 4 on of the 4 cheapest hours":   {hourPrice: set1, hour: 4, hours: 4, expectedResult: false},
		"Set1: Is hour 23 on of the 24 cheapest hours": {hourPrice: set1, hour: 23, hours: 24, expectedResult: true},
		"Set1: Is hour 23 on of the 23 cheapest hours": {hourPrice: set1, hour: 23, hours: 23, expectedResult: false},

		"Set2: Is hour 0 the cheapest hour":           {hourPrice: set2, hour: 0, hours: 1, expectedResult: true},
		"Set2: Is hour 1 the cheapest hour":           {hourPrice: set2, hour: 1, hours: 1, expectedResult: false},
		"Set2: Is hour 10 on of the 5 cheapest hours": {hourPrice: set2, hour: 10, hours: 5, expectedResult: false},
	}

	for k, tc := range cases {
		s.HourPrice[time.Now().Format(DateLayout)] = tc.hourPrice
		result := IsCheapestHour(tc.hour, s.CheapestHours(tc.hours))
		if result != tc.expectedResult {
			t.Fatalf("%s: IsCheapestHour\ngot:  %v\nwant: %v\n", k, result, tc.expectedResult)
		}
	}
}
