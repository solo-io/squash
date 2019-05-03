# calls a go process from a shell script, making it the second PID
FROM alpine
WORKDIR /app
ADD _output/sample_app /app
ADD call_app.sh /app
ENTRYPOINT ./call_app.sh
