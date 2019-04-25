## How to run the end to end tests
```bash
BUILD_ID=desired-image-tag make docker-push
ginkgo -r -v .
```

### Override container values used in tests
- The container repo is set from the `solo-project.yaml` file in the root directory.
  - To change the value, change the `test_container_registry` values to your liking.
  - Refer to the [build tool's repo](https://github.com/solo-io/build) for examples of other configurations.
- The image tag is set by the environment variable `BUILD_ID`.
  - This must be set for the build to succeed. You can use any value.
