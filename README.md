# kubernetes-grafana-controller

This controller will maintain the state of a Grafana instance by syncing it with CRDs created in Kubernetes.  It is under active development.

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

### Dashboards

```
apiVersion: samplecontroller.k8s.io/v1alpha1
kind: GrafanaDashboard
metadata:
  name: test-dash
spec:
  dashboardJson: <dashboard json as string>
```

## Tests

### Integration

To run integration tests run `sudo go test` in the `./test` directory.  This is currently a horrible  combination of go tests and Bash scripting.  I'm not really liking the go test framework for integration tests and will probably swap to Python or Bash.

### Unit

Unit tests are currently broken.  They are a leftover from the kubernetes sample controller project and need to be updated to match the existing controller.

## TODO

- Testing
  - [ ] Controller tests
  - [ ] Integration Tests
    - Dashboards
      - [x] Add
      - [x] Update
      - [x] Delete
    - Notification channels
      - [ ] Add
      - [ ] Update
      - [ ] Delete
    - Datasources
      - [ ] Add
      - [ ] Update
      - [ ] Delete
- Support
  - Dashboards
    - [x] Add
    - [x] Update
    - [x] Delete
  - Notification channels
    - [ ] Add
    - [ ] Update
    - [ ] Delete
  - Datasources
    - [ ] Add
    - [ ] Update
    - [ ] Delete
- rename "SampleController" crap everywhere