# Pre-published Documentation



## Use GDB from `squashctl`
### Deploy a sample app
```bash
kubectl apply -f contrib/language/cpp/demo.yaml
```

### Debug with `squashctl`

```bash
> squashctl
# choose gdb
# choose namespace
# choose pod
# confirm
# (wait)
gdb debug port available on local port 54891.
```

Now open `gdb` and tell it to debug the specified port
```bash
gdb
# tell it connect to the desired port
target remote localhost:54891
# create a breakpoint
b main # confirm
# continue execution
c 
# stop execution with control-c
```
