package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mindprince/gonvml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "nvidia"
)

var (
	averageDuration = 10 * time.Second

	up = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "NVML Metric Collection Operational",
		},
	)

	info = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "driver_info",
			Help:      "NVML Info",
		},
		[]string{"version"},
	)

	deviceCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_count",
			Help:      "Count of found nvidia devices",
		},
	)

	deviceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "info",
			Help:      "Info as reported by the device",
		},
		[]string{"index", "minor", "uuid", "name"},
	)

	temperatures = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "temperatures",
			Help:      "Temperature as reported by the device",
		},
		[]string{"minor"},
	)

	powerUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "power_usage",
			Help:      "Power usage as reported by the device",
		},
		[]string{"minor"},
	)

	powerUsageAverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "power_usage_average",
			Help:      "Power usage as reported by the device averaged over 10s",
		},
		[]string{"minor"},
	)

	fanSpeed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fanspeed",
			Help:      "Fan speed as reported by the device",
		},
		[]string{"minor"},
	)

	memoryTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_total",
			Help:      "Total memory as reported by the device",
		},
		[]string{"minor"},
	)

	memoryUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_used",
			Help:      "Used memory as reported by the device",
		},
		[]string{"minor"},
	)

	utilizationMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "utilization_memory",
			Help:      "Memory Utilization as reported by the device",
		},
		[]string{"minor"},
	)

	utilizationGPU = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "utilization_gpu",
			Help:      "GPU utilization as reported by the device",
		},
		[]string{"minor"},
	)

	utilizationGPUAverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "utilization_gpu_average",
			Help:      "Used memory as reported by the device averraged over 10s",
		},
		[]string{"minor"},
	)
)

type Exporter struct {
	up                    prometheus.Gauge
	info                  prometheus.GaugeVec
	deviceCount           prometheus.Gauge
	temperatures          prometheus.GaugeVec
	deviceInfo            prometheus.GaugeVec
	powerUsage            prometheus.GaugeVec
	powerUsageAverage     prometheus.GaugeVec
	fanSpeed              prometheus.GaugeVec
	memoryTotal           prometheus.GaugeVec
	memoryUsed            prometheus.GaugeVec
	utilizationMemory     prometheus.GaugeVec
	utilizationGPU        prometheus.GaugeVec
	utilizationGPUAverage prometheus.GaugeVec
}

type Metrics struct {
	Version string
	Devices []*Device
}

type Device struct {
	Index                 string
	MinorNumber           string
	Name                  string
	UUID                  string
	Temperature           float64
	PowerUsage            float64
	PowerUsageAverage     float64
	FanSpeed              float64
	MemoryTotal           float64
	MemoryUsed            float64
	UtilizationMemory     float64
	UtilizationGPU        float64
	UtilizationGPUAverage float64
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9401", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	prometheus.MustRegister(NewExporter())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>NVML Exporter</title></head>
             <body>
             <h1>NVML Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	fmt.Println("Starting HTTP server on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func NewExporter() *Exporter {
	return &Exporter{
		up:                    up,
		info:                  *info,
		deviceCount:           deviceCount,
		deviceInfo:            *deviceInfo,
		temperatures:          *temperatures,
		powerUsage:            *powerUsage,
		powerUsageAverage:     *powerUsageAverage,
		fanSpeed:              *fanSpeed,
		memoryTotal:           *memoryTotal,
		memoryUsed:            *memoryUsed,
		utilizationMemory:     *utilizationMemory,
		utilizationGPU:        *utilizationGPU,
		utilizationGPUAverage: *utilizationGPUAverage,
	}
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	data, err := collectMetrics()
	if err != nil {
		log.Printf("Failed to collect metrics: %s\n", err)
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
		e.fanSpeed.WithLabelValues(d.MinorNumber).Set(d.FanSpeed)
		e.memoryTotal.WithLabelValues(d.MinorNumber).Set(d.MemoryTotal)
		e.memoryUsed.WithLabelValues(d.MinorNumber).Set(d.MemoryUsed)
		e.powerUsage.WithLabelValues(d.MinorNumber).Set(d.PowerUsage)
		e.powerUsageAverage.WithLabelValues(d.MinorNumber).Set(d.PowerUsageAverage)
		e.temperatures.WithLabelValues(d.MinorNumber).Set(d.Temperature)
		e.utilizationGPU.WithLabelValues(d.MinorNumber).Set(d.UtilizationGPU)
		e.utilizationGPUAverage.WithLabelValues(d.MinorNumber).Set(d.UtilizationGPUAverage)
		e.utilizationMemory.WithLabelValues(d.MinorNumber).Set(d.UtilizationMemory)
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

func collectMetrics() (*Metrics, error) {
	if err := gonvml.Initialize(); err != nil {
		return nil, err
	}
	defer gonvml.Shutdown()

	version, err := gonvml.SystemDriverVersion()
	if err != nil {
		return nil, err
	}

	metrics := &Metrics{
		Version: version,
	}

	numDevices, err := gonvml.DeviceCount()
	if err != nil {
		return nil, err
	}

	for index := 0; index < int(numDevices); index++ {
		device, err := gonvml.DeviceHandleByIndex(uint(index))
		if err != nil {
			return nil, err
		}

		uuid, err := device.UUID()
		if err != nil {
			return nil, err
		}

		name, err := device.Name()
		if err != nil {
			return nil, err
		}

		minorNumber, err := device.MinorNumber()
		if err != nil {
			return nil, err
		}

		temperature, err := device.Temperature()
		if err != nil {
			return nil, err
		}

		powerUsage, err := device.PowerUsage()
		if err != nil {
			return nil, err
		}

		powerUsageAverage, err := device.AveragePowerUsage(averageDuration)
		if err != nil {
			return nil, err
		}

		fanSpeed, err := device.FanSpeed()
		if err != nil {
			return nil, err
		}

		memoryTotal, memoryUsed, err := device.MemoryInfo()
		if err != nil {
			return nil, err
		}

		utilizationGPU, utilizationMemory, err := device.UtilizationRates()
		if err != nil {
			return nil, err
		}

		utilizationGPUAverage, err := device.AverageGPUUtilization(averageDuration)
		if err != nil {
			return nil, err
		}

		metrics.Devices = append(metrics.Devices,
			&Device{
				Index:                 strconv.Itoa(index),
				MinorNumber:           strconv.Itoa(int(minorNumber)),
				Name:                  name,
				UUID:                  uuid,
				Temperature:           float64(temperature),
				PowerUsage:            float64(powerUsage),
				PowerUsageAverage:     float64(powerUsageAverage),
				FanSpeed:              float64(fanSpeed),
				MemoryTotal:           float64(memoryTotal),
				MemoryUsed:            float64(memoryUsed),
				UtilizationMemory:     float64(utilizationMemory),
				UtilizationGPU:        float64(utilizationGPU),
				UtilizationGPUAverage: float64(utilizationGPUAverage),
			})
	}

	return metrics, nil
}
