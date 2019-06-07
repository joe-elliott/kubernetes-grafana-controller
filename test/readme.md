
# Tests

## Integration

To run integration tests navigate to the `./test` directory and run:

- `./one_time_setup.sh`
  - This sets up minikube and other supporting configuration.  WARNING: it will destroy your existing minikube setup.
- `bats integration_test.bats` 
  - Run this as many time as you want while iterating on the tests.  Note that code changes require re-running `one_time_setup.sh`.
- `./one_time_teardown.sh` 
  - Stops and deletes the minikube cluster 

### Results

With a little bit of luck after running `bats integration_test.bats` you will see

```
$ bats integration_test.bats 
 ✓ creating a Dashboard object creates a Grafana Dashboard
 ✓ deleting a Dashboard object deletes the Grafana Dashboard
 ✓ deleting a Dashboard while the controller is not running deletes the dashboard in Grafana
 ✓ creating a Dashboard object creates the same dashboard in Grafana
 ✓ updating a Dashboard object updates the dashboard in Grafana
 ✓ state is resynced after deleting a dashboard in grafana
 ✓ creating a AlertNotification object creates a Grafana Alert Notification
 ✓ deleting a AlertNotification object deletes the Grafana AlertNotification
 ✓ deleting a AlertNotification while the controller is not running deletes the alert notification in Grafana
 ✓ creating a AlertNotification object creates the same notification in Grafana
 ✓ updating a AlertNotification object updates the notification in Grafana
 ✓ state is resynced after deleting an alert notification in grafana
 ✓ creating a DataSource object creates a Grafana DataSource
 ✓ deleting a DataSource object deletes the Grafana DataSource
 ✓ deleting a DataSource while the controller is not running deletes the datasource in Grafana
 ✓ creating a DataSource object creates the same datasource in Grafana
 ✓ updating a DataSource object updates the datasource in Grafana
 ✓ state is resynced after deleting a datasoure in grafana
```

### Dependencies

- minikube
- kubectl
- docker
- bats (https://github.com/bats-core/bats-core)
- jq
  