# calls the go process only, making it the first PID
FROM alpine
WORKDIR /app
ADD _output/sample_app /app
ENTRYPOINT ./sample_app
