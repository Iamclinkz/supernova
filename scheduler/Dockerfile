FROM alpine:latest
COPY build/dockerBin/scheduler /app/scheduler/
WORKDIR /app/scheduler
EXPOSE 8080

CMD ["./scheduler", \
     "-httpPort=8080", \
     "-logLevel=0", \
     "-discoveryType=k8s", \
     "-dbType=mysql", \
     "-mysqlHost=9.134.5.191", \
     "-mysqlPort=3306", \
     "-mysqlUsername=root", \
     "-mysqlPassword=password",  \
     "-mysqlDbname=supernova",\
     "-mysqlMaxIdleConn=16", \
     "-mysqlMaxOpenConn=128"]