find which container we need to debug
if we have skaffold use that
if not show a menu to chose from

now we have pod. detect language.
either find it from current dir or from pod

we have a language now match it with debugger.
start debugger container with parameters:
namespace and name of pod.
the container will find the pid to debug like in squash
then will exec to dlv

the command line will attach to it.