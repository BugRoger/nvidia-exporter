package main

import (
	"strconv"
	"time"

	"github.com/mindprince/gonvml"
)

var (
	averageDuration = 10 * time.Second
)

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

func collectMetrics() (*Metrics, error) {

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
