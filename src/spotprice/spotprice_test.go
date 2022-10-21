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
		t.Errorf("init() succeeded when it should have failed")
	}

	os.Setenv("TOKEN", "12345")
	if err := s.Init(); err != nil {
		t.Errorf("init() did not succeed")
	}
	os.Unsetenv("TOKEN")
}
