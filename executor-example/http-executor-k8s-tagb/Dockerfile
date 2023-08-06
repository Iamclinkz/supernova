FROM alpine:latest

COPY build/dockerBin/http-executor /app/http-executor/

WORKDIR /app/http-executor

ENV env=k8s

EXPOSE 7070

CMD ["./http-executor", "-grpcPort=7070", \
     "-logLevel=2", \
     "-k8sHealthCheckPort=9090", \
     "-k8sNamespace=supernova"]