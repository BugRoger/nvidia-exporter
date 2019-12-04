package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const namespace = "nvidia"

type Exporter struct {
	up                    prometheus.Gauge
	info                  *prometheus.GaugeVec
	deviceCount           prometheus.Gauge
	temperatures          *prometheus.GaugeVec
	deviceInfo            *prometheus.GaugeVec
	powerUsage            *prometheus.GaugeVec
	powerUsageAverage     *prometheus.GaugeVec
	fanSpeed              *prometheus.GaugeVec
	memoryTotal           *prometheus.GaugeVec
	memoryUsed            *prometheus.GaugeVec
	utilizationMemory     *prometheus.GaugeVec
	utilizationGPU        *prometheus.GaugeVec
	utilizationGPUAverage *prometheus.GaugeVec
}

func main() {
	var (
		level         = flag.String("log.level", "info", "Set the output log level")
		listenAddress = flag.String("web.listen-address", "0.0.0.0:9401", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()
	setLogLevel(*level)

	prometheus.MustRegister(NewExporter())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>NVML Exporter</title></head>
             <body>
             <h1>NVML Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
	     <h2>More information:</h2>
	     <p><a href="https://github.com/BugRoger/nvidia-exporter">github.com/BugRoger/nvidia-exporter</a></p>
             </body>
             </html>`))
	})
	log.Infof("Starting HTTP server on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func setLogLevel(level string) {
	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized minimum log level; using 'info' as default")
	}
}

func NewExporter() *Exporter {
	return &Exporter{
		up: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "NVML Metric Collection Operational",
			},
		),
		info: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "driver_info",
				Help:      "NVML Info",
			},
			[]string{"version"},
		),
		deviceCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "device_count",
				Help:      "Count of found nvidia devices",
			},
		),
		deviceInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "info",
				Help:      "Info as reported by the device",
			},
			[]string{"index", "minor", "uuid", "name"},
		),
		temperatures: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "temperatures",
				Help:      "Temperature as reported by the device",
			},
			[]string{"minor"},
		),
		powerUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "power_usage",
				Help:      "Power usage as reported by the device",
			},
			[]string{"minor"},
		),
		powerUsageAverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "power_usage_average",
				Help:      "Power usage as reported by the device averaged over 10s",
			},
			[]string{"minor"},
		),
		fanSpeed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "fanspeed",
				Help:      "Fan speed as reported by the device",
			},
			[]string{"minor"},
		),
		memoryTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_total",
				Help:      "Total memory as reported by the device",
			},
			[]string{"minor"},
		),
		memoryUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_used",
				Help:      "Used memory as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_memory",
				Help:      "Memory Utilization as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationGPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_gpu",
				Help:      "GPU utilization as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationGPUAverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_gpu_average",
				Help:      "Used memory as reported by the device averraged over 10s",
			},
			[]string{"minor"},
		),
	}
}

// This function is used to check if metric
// value is valid; we expect nothing less than 0
// gonvml returns uint data type
func checkMetric(value float64) bool {
	if value < 0 {
		return false
	} else {
		return true
	}
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	data, err := collectMetrics()
	if err != nil {
		log.Errorf("Failed to collect metrics: %s", err)
		e.up.Set(0)
		e.up.Collect(metrics)
		return
	}

	e.up.Set(1)
	e.info.WithLabelValues(data.Version).Set(1)
	e.deviceCount.Set(float64(len(data.Devices)))

	for i := 0; i < len(data.Devices); i++ {
		d := data.Devices[i]
		e.deviceInfo.WithLabelValues(d.Index, d.MinorNumber, d.Name, d.UUID).Set(1)
		if checkMetric(d.FanSpeed) {
			e.fanSpeed.WithLabelValues(d.MinorNumber).Set(d.FanSpeed)
		}
		if checkMetric(d.MemoryTotal) {
			e.memoryTotal.WithLabelValues(d.MinorNumber).Set(d.MemoryTotal)
		}
		if checkMetric(d.MemoryUsed) {
			e.memoryUsed.WithLabelValues(d.MinorNumber).Set(d.MemoryUsed)
		}
		if checkMetric(d.PowerUsage) {
			e.powerUsage.WithLabelValues(d.MinorNumber).Set(d.PowerUsage)
		}
		if checkMetric(d.PowerUsageAverage) {
			e.powerUsageAverage.WithLabelValues(d.MinorNumber).Set(d.PowerUsageAverage)
		}
		if checkMetric(d.Temperature) {
			e.temperatures.WithLabelValues(d.MinorNumber).Set(d.Temperature)
		}
		if checkMetric(d.UtilizationGPU) {
			e.utilizationGPU.WithLabelValues(d.MinorNumber).Set(d.UtilizationGPU)
		}
		if checkMetric(d.UtilizationGPUAverage) {
			e.utilizationGPUAverage.WithLabelValues(d.MinorNumber).Set(d.UtilizationGPUAverage)
		}
		if checkMetric(d.UtilizationMemory) {
			e.utilizationMemory.WithLabelValues(d.MinorNumber).Set(d.UtilizationMemory)
		}
	}

	e.deviceCount.Collect(metrics)
	e.deviceInfo.Collect(metrics)
	e.fanSpeed.Collect(metrics)
	e.info.Collect(metrics)
	e.memoryTotal.Collect(metrics)
	e.memoryUsed.Collect(metrics)
	e.powerUsage.Collect(metrics)
	e.powerUsageAverage.Collect(metrics)
	e.temperatures.Collect(metrics)
	e.up.Collect(metrics)
	e.utilizationGPU.Collect(metrics)
	e.utilizationGPUAverage.Collect(metrics)
	e.utilizationMemory.Collect(metrics)
}

func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	e.deviceCount.Describe(descs)
	e.deviceInfo.Describe(descs)
	e.fanSpeed.Describe(descs)
	e.info.Describe(descs)
	e.memoryTotal.Describe(descs)
	e.memoryUsed.Describe(descs)
	e.powerUsage.Describe(descs)
	e.powerUsageAverage.Describe(descs)
	e.temperatures.Describe(descs)
	e.up.Describe(descs)
	e.utilizationGPU.Describe(descs)
	e.utilizationGPUAverage.Describe(descs)
	e.utilizationMemory.Describe(descs)
}
