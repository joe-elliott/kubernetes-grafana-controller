# kubernetes-grafana-controller

[![Go Report Card](https://goreportcard.com/badge/github.com/number101010/kubernetes-grafana-controller)](https://goreportcard.com/report/github.com/number101010/kubernetes-grafana-controller)

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
apiVersion: grafana.k8s.io/v1alpha1
kind: GrafanaDashboard
metadata:
  name: test-dash
spec:
  dashboardJson: <dashboard json as string>
```

## Tests

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

### Unit

Unit tests for the controller technically pass, but they only test creating a new dashboard.  We need to flesh out grafana client testing structure and increase coverage.

## TODO

- Testing
  - [ ] Controller tests
    - Define and build tests
  - [ ] Integration Tests
    - Inspect events and pods logs as part of testing
    - Reduce dependencies by running bats in container
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
    - [x] Add
    - [x] Update
    - [x] Delete
      - Currently using name as the primary key/id for the notification channel.  Determine/document appropriate behavior when multiple channels have the same names.
  - Datasources
    - [x] Add
    - [x] Update
    - [x] Delete
- Refactoring/Cleanup/Additional
  - Add comments where go wants me to
  - The dashboard object is currently `grafanadashboard.grafana.k8s.io`.  This feels wrong.  Revisit object naming.  Should it be `dashboard.grafana.com`, `dashboard.kubernetes-grafana-controller`?
  - Add prometheus metrics
  - Determine/document the behavior of a dashboard with a uid vs one without.  Confirm sanity.
     - use uid?  use id?  just use name as primary key?  need to investigate various options and pick a path
     - Since Grafana wants to use the name as a unique key should we stop storing IDs in status and just query the id using the name when we want to delete/update?
  - Fix API objects naming to be less verbose and redundant
    - Drop "Grafana" on all objects
    - Change "Notification Channels" to "Alert Notifications"
  - Support interpolation of k8s secrets into datasources
  - Grafana client is a mess.  clean it up
  - leaving version in a grafana object can cause failed updates.  remove?
  - the logic to prevent re-updates is terrible.  check actual state of grafana?
  - confirm 30 second resync works
  - confirm appropriate use of utilruntime.HandleError vs. return err
  - improve logging considerably
  - Tests
    - test that tests to see if the controller will eventually sync even if grafana is down
    - consolidate integration test logic: i.e. all post tests are basically the same
    - create more focused tests.  each test tests too many things
  - create issues from this list ^
