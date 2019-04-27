# About
This is a very simple app that prints a new line every second for 1000 seconds.

Your challenge is to attach a debugger before the time runs out.

# Deploy the app
```bash
kubectl apply -f demo.yaml
```

# To build and run locally
```bash
make local
./_out/squash-demo-cpp
```

# To deploy a modified app
```bash
# make changes to the app
export VERSION=some-version
make build-push -B
# edit the demo.yaml to reflect your new version
kubectl apply -f demo.yaml
```

how to use gdb:
https://www.cs.umd.edu/~srhuang/teaching/cmsc212/gdb-tutorial-handout.pdf

Getting permission for gdb on mac:
https://sourceware.org/gdb/wiki/PermissionsDarwin#Create_a_certificate_in_the_System_Keychain

