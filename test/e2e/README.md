## How to run the end to end tests
```bash
make docker-push
ginkgo -r -v .
```

### Override container values used in tests
- By default, the test will choose which plank container repo and image tag by parsing the `_output/buildtimevalues.yaml` file produced during build.
- This ensures that the test will run the expected containers.
- **To override** set the following environment values:
```bash
export PLANK_IMAGE_TAG=<preferred_image_tag>
export PLANK_IMAGE_REPO=<preferred_repo>
```
