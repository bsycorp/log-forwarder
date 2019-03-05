package main

import (
	metrics "github.com/rcrowley/go-metrics"
	"time"
)

type Metrics struct {
	Registry         metrics.Registry
	NumActiveBuffers metrics.Gauge
	MainLoopSpins    metrics.Counter
	UploadSuccess    metrics.Counter
	UploadFailure    metrics.Counter
}

func (m *Metrics) Init() {
	m.Registry = metrics.NewRegistry()

	metrics.RegisterDebugGCStats(m.Registry)
	metrics.RegisterRuntimeMemStats(m.Registry)

	m.NumActiveBuffers = metrics.NewGauge()
	m.UploadSuccess = metrics.NewCounter()
	m.UploadFailure = metrics.NewCounter()
	m.MainLoopSpins = metrics.NewCounter()

	_ = metrics.Register("num_active_buffers.gauge", m.NumActiveBuffers)
	_ = metrics.Register("upload.success.count", m.UploadSuccess)
	_ = metrics.Register("upload.failure.count", m.UploadSuccess)
	_ = metrics.Register("main_loop_spins.count", m.MainLoopSpins)
}

func (m *Metrics) StartCapture() {
	go metrics.CaptureDebugGCStats(m.Registry, time.Second*5)
	go metrics.CaptureRuntimeMemStats(m.Registry, time.Second*5)
}
