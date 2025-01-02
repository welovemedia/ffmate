package sev

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	promDto "github.com/prometheus/client_model/go"
)

type Stats struct {
	AppName    string `json:"appName"`
	AppVersion string `json:"appVersion"`

	RuntimeDuration time.Duration `json:"runtimeDuration"`

	Os   string `json:"os"`
	Arch string `json:"arch"`

	Metrics map[string]float64

	Custom map[string]interface{}
}

func (s *Sev) SendTelemtry(targetUrl string, custom map[string]interface{}) {
	stats := Stats{
		AppName:    s.AppName(),
		AppVersion: s.AppVersion(),

		RuntimeDuration: time.Since(s.AppStartTime()),

		Os:   runtime.GOOS,
		Arch: runtime.GOARCH,

		Metrics: make(map[string]float64),

		Custom: custom,
	}

	for name, gauge := range s.Metrics().Gauges() {
		g := &promDto.Metric{}
		gauge.Write(g)
		stats.Metrics[name] = g.Gauge.GetValue()
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		s.Logger().Warnf("failed to marshal telemetry data: %+v", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(b))
	if err != nil {
		s.Logger().Error("failed to create http request", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", s.AppName()+"/"+s.AppVersion())

	_, err = client.Do(req)
	if err != nil {
		s.Logger().Warnf("failed to send telemtry data: %+v", err)
	} else {
		s.Logger().Debugf("sent telemetry data")
	}
}
