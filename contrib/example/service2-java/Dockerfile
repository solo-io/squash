FROM java:8  
COPY target/service2-1.0-SNAPSHOT-shaded.jar /service2-1.0-SNAPSHOT-shaded.jar 
ENTRYPOINT java -agentlib:jdwp=transport=dt_socket,server=y,address=8000,suspend=n -jar service2-1.0-SNAPSHOT-shaded.jar 

EXPOSE 8080
