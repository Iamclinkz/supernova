FROM alpine:latest
COPY build/scheduler /app/scheduler/
WORKDIR /app/scheduler
EXPOSE 8080

ENV env=k8s \
    httpPort=8080 \
    logLevel=1 \
    discoveryType=k8s \
    dbType=mysql \
    mysqlHost=9.134.5.191 \
    mysqlPort=3306 \
    mysqlUsername=root \
    mysqlPassword=password \
    mysqlDbname=supernova \
    mysqlMaxIdleConn=16 \
    mysqlMaxOpenConn=128

CMD ["./scheduler", \
     "-httpPort", "${httpPort}", \
     "-logLevel", "${logLevel}", \
     "-discoveryType", "${discoveryType}", \
     "-dbType", "${dbType}", \
     "-mysqlHost", "${mysqlHost}", \
     "-mysqlPort", "${mysqlPort}", \
     "-mysqlUsername", "${mysqlUsername}", \
     "-mysqlPassword", "${mysqlPassword}", \
     "-mysqlDbname", "${mysqlDbname}", \
     "-mysqlMaxIdleConn", "${mysqlMaxIdleConn}", \
     "-mysqlMaxOpenConn", "${mysqlMaxOpenConn}"]