# kubernetes-grafana-controller

[![Go Report Card](https://goreportcard.com/badge/github.com/number101010/kubernetes-grafana-controller)](https://goreportcard.com/report/github.com/number101010/kubernetes-grafana-controller)

This controller will maintain the state of a Grafana instance by syncing it with objects created in Kubernetes.  It is under active development.

The primary motivator for creating this controller is to allow development teams to include their Grafana dashboards in the same source repos as their code alongside other Kubernetes objects.

## CLI

```
  -grafana string
    	The address of the Grafana server. (default "http://grafana")
  -kubeconfig string
    	Path to a kubeconfig. Only required if out-of-cluster.
  -master string
    	The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.

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

## Custom Resource Definitions

The kubernetes-grafana-controller currently will sync the following objects.

### Dashboards

```
apiVersion: grafana.k8s.io/v1alpha1
kind: GrafanaDashboard
metadata:
  name: test
spec:
  dashboardJson: <dashboard json as string>
```

### Notification Channels

```
apiVersion: grafana.k8s.io/v1alpha1
kind: GrafanaNotificationChannel
metadata:
  name: test
spec:
  notificationChannelJson: <channel json as string>
```

### Datasources

```
apiVersion: grafana.k8s.io/v1alpha1
kind: GrafanaDataSource
metadata:
  name: test
spec:
  dataSourceJson: <data source json as string>
```

## Tests

[tests](test/readme.md)

### Integration

To run integration tests navigate to the `./test` directory and run:

- `./one_time_setup.sh`
  - This sets up minikube and other supporting configuration
- `bats integration_test.bats` 
  - Run this as many time as you want while iterating on the tests.  Note that code changes require re-running `one_time_setup.bats`.
- `./one_time_teardown.sh` 
  - Stops and deletes the minikube cluster 

Previously I had attempted to use the native go testing framework for integration tests.  However, since the tests were basically a long list of bash commands it made for some super gross code.  Moving to bats simplified and improved the integration tests.

#### Dependencies

- minikube
- kubectl
- docker
- bats (https://github.com/bats-core/bats-core)
- jq
  