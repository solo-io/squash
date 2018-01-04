
# Java debugger

Java process must be started with [JDWP](http://docs.oracle.com/javase/7/docs/technotes/guides/jpda/jdwp-spec.html) options.
for example: *java -Xdebug -Xrunjdwp:server=y,transport=dt_socket,address=8000,suspend=n -jar MyApp.jar*
