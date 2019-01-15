

# Dev workflow notes

## setup a watcher to inspect the debug resources
```
cd test/dev/watcher
go run main
```

## initialize some sample apps and the squash client
```
cd test/dev
go run main --init # to load sample apps and squash client
go run main --att # make an attachment
go run main --clean # remove resources

# whenever you make changes to the squash client (after rebuilding)
go run main --init && go run main --clean
```

## run the e2e tests
```
cd test/e2e
ginkgo -r
```
