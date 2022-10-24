package control

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	s := State{}

	if err := s.Init(false); err != nil {
		t.Errorf("init() with defaults did not succeed")
	}

	if err := s.Init(true); err != nil {
		t.Errorf("init() with defaults did not succeed")
	}

	os.Setenv("SHELLY_URL", "http://127.0.0.1")
	if err := s.Init(false); err != nil {
		t.Errorf("init() did not succeed")
	}
	os.Unsetenv("SHELLY_URL")
}

// Need to mock http client..
//func TestSwitchOn(t *testing.T) {
//	cases := map[string]struct {
//		expectedResult error
//	}{
//		"SwitchON when Switch is OFF": {
//			expectedResult: nil,
//		},
//	}
//
//	os.Setenv("SHELLY_URL", "http://")
//	s := State{}
//	s.Init()
//
//	for k, tc := range cases {
//		err := s.SwitchOn()
//		if err != tc.expectedResult {
//			t.Fatalf("%s: SwtichOn\ngot:  %s\nwant: %s\n", k, err, tc.expectedResult)
//		}
//	}
//}
