
# Tests

## Integration

To run integration tests navigate to the `./test` directory and run:

- `./one_time_setup.sh`
  - This sets up minikube and other supporting configuration.  WARNING: it will destroy your existing minikube setup.
- `bats integration_test.bats` 
  - Run this as many time as you want while iterating on the tests.  Note that code changes require re-running `one_time_setup.sh`.
- `./one_time_teardown.sh` 
  - Stops and deletes the minikube cluster 

Previously I had attempted to use the native go testing framework for integration tests.  However, since the tests were basically a long list of bash commands it made for some super gross code.  Moving to bats simplified and improved the integration tests.

### Dependencies

- minikube
- kubectl
- docker
- bats (https://github.com/bats-core/bats-core)
- jq
  