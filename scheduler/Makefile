RUN_NAME := scheduler
IMAGE_NAME := exiasun/scheduler
TAG := latest
DEPLOYMENT_NAME := scheduler-service
YAML := "../deployment/scheduler/scheduler.yaml"

build-binary:
	@echo "--------------------------------------Building ${RUN_NAME} binary..."
	go build -o build/bin/${RUN_NAME} *.go

image-runnable:
	@echo "--------------------------------------Building ${RUN_NAME} binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/dockerBin/${RUN_NAME} *.go

docker-build:
	@echo "--------------------------------------Building docker image ${IMAGE_NAME}:${TAG}"
	docker build -t ${IMAGE_NAME}:${TAG} .
	- @docker rmi $(docker images --filter "dangling=true" -q)

minikube-load:
	@echo "--------------------------------------Load image to minikube: ${IMAGE_NAME}:${TAG}"
	- minikube kubectl -- delete -f ${YAML}
# sleep 20
# - minikube ssh docker rmi ${IMAGE_NAME}:${TAG}
# minikube image load ${IMAGE_NAME}:${TAG} --overwrite=true
# minikube kubectl -- apply -f ${YAML}

all: build-binary

k8s: image-runnable docker-build minikube-load

clean:
	rm -rf build/*