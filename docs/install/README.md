# Installing Squash

To install Squash you need to install the Squash server and clients on your container orchestration platform of choice, and to install the CLI on your computer. 

### Platforms
Currently Suqash only supports Kubernets. Other platforms coming up... 
 - [Kubernetes](kubernetes.md)

### Command Line Interface (CLI)
Download the CLI binary:
- [Linux](link)     
- [OS X](link)

```
curl ...
chmod +x squash
```
The easiest way is to place it somewhere in your path, but it is not a must.

- Make sure you have access to the squash server - use `$ squash app list` to test that it is working properly.
- Make sure kubectl port-foward functionality works.

If you have an issue with either, see the [FAQ](faq.md) for help.

