package spotprice

import (
	"os"
	"testing"
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
