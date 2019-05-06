# calls the go process only, making it the first PID
FROM alpine
WORKDIR /app
ADD _output/log_spooler_ping /app
EXPOSE 8080
ENTRYPOINT ./log_spooler_ping
