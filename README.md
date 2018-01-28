# NVIDIA Prometheus Exporter

This exporter used the NVidia Management Library (NVML) to query information
about the installed Nvidia GPUs.

## Requirements

The NVML shared library (libnvidia-ml.so.1) need to be loadable. When running
in a container it must be either baked in or mounted from the host.

## Running in Kubernetes

```
apiVersion: apps/v1beta1 
kind: Deployment

metadata:
  name: nvidia-exporter 
spec:
  replicas: 1 
  strategy: 
    type: Recreate
  template:
    metadata:
      labels:
        app: nvidia-exporter
    spec:
      containers:
        - name: nvidia-exporter
          securityContext:
            privileged: true
          image: bugroger/nvidia-exporter:latest
          ports:
            - containerPort: 9401 
          volumeMounts:
            - mountPath: /usr/local/nvidia
              name: nvidia 
      volumes:
        - name: nvidia
          hostPath:
            path: /opt/nvidia/current
---
apiVersion: v1
kind: Service

metadata:
  name: nvidia-exporter 
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9401"
spec:
  selector:
    app: nvidia-exporter
  ports:
    - name: tcp
      port: 9401
```

## Example

```
# HELP nvidia_device_count Count of found nvidia devices
# TYPE nvidia_device_count gauge
nvidia_device_count 6
# HELP nvidia_driver_info NVML Info
# TYPE nvidia_driver_info gauge
nvidia_driver_info{version="384.111"} 1
# HELP nvidia_fanspeed Fan speed as reported by the device
# TYPE nvidia_fanspeed gauge
nvidia_fanspeed{minor="0"} 40
nvidia_fanspeed{minor="1"} 32
nvidia_fanspeed{minor="2"} 27
nvidia_fanspeed{minor="3"} 42
nvidia_fanspeed{minor="4"} 39
nvidia_fanspeed{minor="5"} 43
# HELP nvidia_info Info as reported by the device
# TYPE nvidia_info gauge
nvidia_info{index="0",minor="0",name="GPU-352c2b3d-5783-6e52-25b7-bc6a9fdb78bb",uuid="GeForce GTX 1070"} 1
nvidia_info{index="1",minor="1",name="GPU-7484d1b6-8b71-15dd-ddda-4dcd3e0d22c6",uuid="GeForce GTX 1070"} 1
nvidia_info{index="2",minor="2",name="GPU-7a604372-6642-db36-be4a-81e9e4d2de59",uuid="GeForce GTX 1070"} 1
nvidia_info{index="3",minor="3",name="GPU-727f0f85-cbfa-c75c-484b-5cd5a71175ba",uuid="GeForce GTX 1070"} 1
nvidia_info{index="4",minor="4",name="GPU-34891dbe-e41c-2568-8af5-84b170805eaf",uuid="GeForce GTX 1070"} 1
nvidia_info{index="5",minor="5",name="GPU-2e31969f-b354-9675-c034-1fce9073951c",uuid="GeForce GTX 1070"} 1
# HELP nvidia_memory_total Total memory as reported by the device
# TYPE nvidia_memory_total gauge
nvidia_memory_total{minor="0"} 8.506048512e+09
nvidia_memory_total{minor="1"} 8.508145664e+09
nvidia_memory_total{minor="2"} 8.508145664e+09
nvidia_memory_total{minor="3"} 8.508145664e+09
nvidia_memory_total{minor="4"} 8.508145664e+09
nvidia_memory_total{minor="5"} 8.508145664e+09
# HELP nvidia_memory_used Used memory as reported by the device
# TYPE nvidia_memory_used gauge
nvidia_memory_used{minor="0"} 5.53517056e+08
nvidia_memory_used{minor="1"} 5.53517056e+08
nvidia_memory_used{minor="2"} 5.53517056e+08
nvidia_memory_used{minor="3"} 5.53517056e+08
nvidia_memory_used{minor="4"} 5.53517056e+08
nvidia_memory_used{minor="5"} 5.53517056e+08
# HELP nvidia_power_usage Power usage as reported by the device
# TYPE nvidia_power_usage gauge
nvidia_power_usage{minor="0"} 98510
nvidia_power_usage{minor="1"} 99647
nvidia_power_usage{minor="2"} 98112
nvidia_power_usage{minor="3"} 97347
nvidia_power_usage{minor="4"} 101280
nvidia_power_usage{minor="5"} 98777
# HELP nvidia_power_usage_average Power usage as reported by the device averaged over 10s
# TYPE nvidia_power_usage_average gauge
nvidia_power_usage_average{minor="0"} 99466
nvidia_power_usage_average{minor="1"} 99373
nvidia_power_usage_average{minor="2"} 99513
nvidia_power_usage_average{minor="3"} 99927
nvidia_power_usage_average{minor="4"} 99611
nvidia_power_usage_average{minor="5"} 99653
# HELP nvidia_temperatures Temperature as reported by the device
# TYPE nvidia_temperatures gauge
nvidia_temperatures{minor="0"} 60
nvidia_temperatures{minor="1"} 55
nvidia_temperatures{minor="2"} 54
nvidia_temperatures{minor="3"} 61
nvidia_temperatures{minor="4"} 59
nvidia_temperatures{minor="5"} 62
# HELP nvidia_up NVML Metric Collection Operational
# TYPE nvidia_up gauge
nvidia_up 1
# HELP nvidia_utilization_gpu GPU utilization as reported by the device
# TYPE nvidia_utilization_gpu gauge
nvidia_utilization_gpu{minor="0"} 100
nvidia_utilization_gpu{minor="1"} 100
nvidia_utilization_gpu{minor="2"} 100
nvidia_utilization_gpu{minor="3"} 100
nvidia_utilization_gpu{minor="4"} 100
nvidia_utilization_gpu{minor="5"} 100
# HELP nvidia_utilization_gpu_average Used memory as reported by the device averraged over 10s
# TYPE nvidia_utilization_gpu_average gauge
nvidia_utilization_gpu_average{minor="0"} 99
nvidia_utilization_gpu_average{minor="1"} 99
nvidia_utilization_gpu_average{minor="2"} 99
nvidia_utilization_gpu_average{minor="3"} 99
nvidia_utilization_gpu_average{minor="4"} 99
nvidia_utilization_gpu_average{minor="5"} 99
# HELP nvidia_utilization_memory Memory Utilization as reported by the device
# TYPE nvidia_utilization_memory gauge
nvidia_utilization_memory{minor="0"} 78
nvidia_utilization_memory{minor="1"} 78
nvidia_utilization_memory{minor="2"} 76
nvidia_utilization_memory{minor="3"} 75
nvidia_utilization_memory{minor="4"} 78
nvidia_utilization_memory{minor="5"} 76
```
