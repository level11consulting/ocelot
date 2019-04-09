package disk_usage

import (
	"testing"
)

func TestHealthChecker_Healthy(t *testing.T) {
	hc := &HealthChecker{
		Path:           "/var/lib",
		PauseThreshold: "1B",
	}
	err := hc.Healthy()
	if err != nil {
		t.Error("it is possible that the machine running this test has less than 1 byte of disk space, and this test failure would then be in error. but it feels unlikely.", err)
	}
	hc.PauseThreshold = "200TB"
	if err = hc.Healthy(); err == nil {
		t.Error("it is possible that the machine running htis test has >200TB of disk space, and this test failure would then be in error. but it feels unlikely.", err)
	}
	hc.PauseThreshold = "100000GB"
	if err = hc.Healthy(); err == nil {
		t.Error("it is possible that the machine running htis test has >100000GB of disk space, and this test failure would then be in error. but it feels unlikely.", err)
	}
	hc.PauseThreshold = "1q"
	if err = hc.Healthy(); err == nil {
		t.Error("invalid threshold, should fail")
	}
	hc.PauseThreshold = ""
	if err = hc.Healthy(); err != nil {
		t.Error("an empty pausethreshold should result in no checks being made and no error")
	}
	hc.PauseThreshold = "1B"
	hc.Path = ""
	if err = hc.Healthy(); err != nil {
		t.Error("an empty path should result in no checks being made and no error")
	}
}
