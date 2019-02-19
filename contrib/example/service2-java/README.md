# Java Service 2

Java implementation of Service 2 with the bug

## Build

Build the deployment Jar file by running

    mvn package

## Run

Run the service with Java by running

    java -jar target/service2-1.0-SNAPSHOT-shaded.jar




## Test

- test with:
```
curl --header "Content-Type: application/json"   --request POST   --data '{"Op1":"4","Op2":5,"IsAdd":"true"}'   http://localhost:8080/calculate
```
