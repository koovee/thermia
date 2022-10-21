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

	if err := s.Init(); err != nil {
		t.Errorf("init() with defaults did not succeed")
	}

	os.Setenv("SHELLY_URL", "http://127.0.0.1")
	if err := s.Init(); err != nil {
		t.Errorf("init() did not succeed")
	}
	os.Unsetenv("SHELLY_URL")
}
