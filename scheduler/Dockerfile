FROM alpine:latest
COPY build/dockerBin/scheduler /app/scheduler/
WORKDIR /app/scheduler
EXPOSE 8080

ENV env=k8s

CMD ["./scheduler", \
     "-httpPort=8080", \
     "-logLevel=2", \
     "-discoveryType=k8s", \
     "-dbType=mysql", \
     "-mysqlHost=9.134.5.191", \
     "-mysqlPort=3306", \
     "-mysqlUsername=root", \
     "-mysqlPassword=password",  \
     "-mysqlDbname=supernovaK8s",\
     "-mysqlMaxIdleConn=16", \
     "-mysqlMaxOpenConn=128"]