package healthcheck

import (
	"github.com/r3boot/anycast-agent/lib"
	"time"
)

var Logger lib.Logger

type HealthCheckConfig struct {
	Command     string
	Interval    time.Duration
	InitDamping int
	MaxRetries  int
	ResultChan  chan bool
}

type HealthCheck struct {
	Config HealthCheckConfig
	Health bool
}

func NewHealthCheck(logger lib.Logger, cfg HealthCheckConfig) *HealthCheck {
	Logger = logger
	return &HealthCheck{Config: cfg}
}

func (hc *HealthCheck) RunCheckRoutine() {
	var (
		checkStatus bool
	)

	Logger.Debug("HealthCheck: Starting polling routine")
	for {
		checkStatus = lib.RunsOK(hc.Config.Command)
		hc.Config.ResultChan <- checkStatus

		time.Sleep(hc.Config.Interval)
	}
}
