# kubernetes-grafana-controller

[![Go Report Card](https://goreportcard.com/badge/github.com/joe-elliott/kubernetes-grafana-controller)](https://goreportcard.com/report/github.com/joe-elliott/kubernetes-grafana-controller)

This controller will maintain the state of a Grafana instance by syncing it with objects created in Kubernetes.  It is under active development.

The primary motivator for creating this controller is to allow development teams to include their Grafana dashboards in the same source repos as their code alongside other Kubernetes objects.

- [tests](test/readme.md)

## CLI

```
  -grafana string
    	The address of the Grafana server. (default "http://grafana")
  -kubeconfig string
    	Path to a kubeconfig. Only required if out-of-cluster.
  -master string
    	The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
  -resync duration
    	Periodic interval in which to force resync objects. (default 30s)
  -resync-delete duration
    	Periodic interval in which to force resync deleted objects.  Pass 0s to disable. (default 30s)
  -prometheus-listen-address string
    	The address to listen on for Prometheus scrapes. (default ":8080")
  -prometheus-path string
    	The path to publish Prometheus metrics to. (default "/metrics")

klog

  -alsologtostderr
    	log to standard error as well as files
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -log_file string
    	If non-empty, use this log file
  -logtostderr
    	log to standard error instead of files
  -skip_headers
    	If true, avoid header prefixes in the log messages
  -stderrthreshold value
    	logs at or above this threshold go to stderr (default 2)
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
```

## Metrics

The kubernetes-grafana-controller publishes a metrics in the prometheus format.  These include error totals, grafana latencies and other totals.

https://github.com/joe-elliott/kubernetes-grafana-controller/blob/master/pkg/prometheus/prometheus.go

## Custom Resource Definitions

The kubernetes-grafana-controller currently will sync the following objects.

### Dashboards

```
apiVersion: grafana.com/v1alpha1
kind: Dashboard
metadata:
  name: test
spec:
  folderName: <optional name of a folder object to place this dashboard in>
  json: <dashboard json as string>
```

### Folders

```
apiVersion: grafana.com/v1alpha1
kind: Folder
metadata:
  name: test
spec:
  json: <folder json as string>
```

### AlertNotifications (Notification Channels)

```
apiVersion: grafana.com/v1alpha1
kind: AlertNotification
metadata:
  name: test
spec:
  json: <notification json as string>
```

### DataSources

```
apiVersion: grafana.com/v1alpha1
kind: DataSource
metadata:
  name: test
spec:
  json: <data source json as string>
```

## Requirements

This controller requires the `CustomResourceSubresources` feature gate to enabled.  This has been enabled by default since k8s 1.11.
