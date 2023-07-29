FROM alpine:latest

COPY build/dockerBin/http-executor /app/http-executor/

WORKDIR /app/http-executor

EXPOSE 7070

CMD ["./http-executor", "-grpcPort=7070", \
     "-logLevel=0", \
     "-k8sHealthCheckPort=9090", \
     "-k8sNamespace=supernova"]