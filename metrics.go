package main

import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/syntaqx/go-metrics-datadog"
	"log"
	"strings"
	"time"
)

type Metrics struct {
	Registry                metrics.Registry
	DebugDupCursor          metrics.Counter
	DebugSkippedCursor      metrics.Counter
	BuffersActive           metrics.Gauge
	BufferUploadSuccess     metrics.Counter
	BufferUploadFailure     metrics.Counter
	MainLoopSpins           metrics.Counter
	MainLoopTime            metrics.Timer
	UploadMessages          metrics.Counter
	UploadBytesUncompressed metrics.Counter
	UploadBytesCompressed   metrics.Counter
	UploadTime              metrics.Timer
}

func (m *Metrics) Init() {
	m.Registry = metrics.NewRegistry()

	metrics.RegisterDebugGCStats(m.Registry)
	metrics.RegisterRuntimeMemStats(m.Registry)

	m.DebugDupCursor = metrics.NewCounter()
	m.DebugSkippedCursor = metrics.NewCounter()
	m.BuffersActive = metrics.NewGauge()
	m.BufferUploadSuccess = metrics.NewCounter()
	m.BufferUploadFailure = metrics.NewCounter()
	m.MainLoopTime = metrics.NewTimer()
	m.MainLoopSpins = metrics.NewCounter()
	m.UploadMessages = metrics.NewCounter()
	m.UploadBytesUncompressed = metrics.NewCounter()
	m.UploadBytesCompressed = metrics.NewCounter()
	m.UploadTime = metrics.NewTimer()

	_ = metrics.Register("debug.dup_cursor.count", m.DebugDupCursor)
	_ = metrics.Register("debug.skipped_cursor.count", m.DebugSkippedCursor)
	_ = metrics.Register("buffers.active.gauge", m.BuffersActive)
	_ = metrics.Register("buffers.upload.success", m.BufferUploadSuccess)
	_ = metrics.Register("buffers.upload.failure", m.BufferUploadFailure)
	_ = metrics.Register("main_loop_spin.count", m.MainLoopSpins)
	_ = metrics.Register("main_loop_spin.time_ms", m.MainLoopTime)
	_ = metrics.Register("upload.messages.count", m.UploadMessages)
	_ = metrics.Register("upload.bytes.uncompressed.count", m.UploadBytesUncompressed)
	_ = metrics.Register("upload.bytes.compressed.count", m.UploadBytesCompressed)
	_ = metrics.Register("upload.time_ms", m.UploadTime)
}

func (m *Metrics) Start(metricsArg string) {
	// Metrics
	if metricsArg == "none" {
		log.Println("metrics reporting disabled")
	} else if strings.HasPrefix(metricsArg, "datadog") {
		dogStatsdAddr := "127.0.0.1:8125"
		bits := strings.Split(metricsArg, ",")
		if len(bits) > 1 {
			dogStatsdAddr = bits[1]
		}
		reporter, err := datadog.NewReporter(
			m.Registry,        // Metrics registry
			dogStatsdAddr,     // DogStatsD UDP address
			time.Second*10,    // Update interval
			datadog.UsePercentiles([]float64{0.25, 0.99}),
		)
		if err != nil {
			log.Fatal(err)
		}
		reporter.Client.Namespace = "log_forwarder."
		go reporter.Flush()
		log.Println("metrics reporting to dogstatd at: ", dogStatsdAddr)
	} else {
		log.Fatal("unknown metrics provider: ", metricsArg)
	}
	go metrics.CaptureRuntimeMemStats(m.Registry, time.Second*5)
	// go metrics.CaptureDebugGCStats(m.Registry, time.Second*5)
}
