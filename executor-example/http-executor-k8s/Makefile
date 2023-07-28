RUN_NAME := http-executor
IMAGE_NAME := exiasun/http-executor
TAG := dev
DEPLOYMENT_NAME := http-executor-service

image-runnable:
	@echo "--------------------------------------Building ${RUN_NAME} binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/${RUN_NAME} *.go

docker-build:
	@echo "--------------------------------------Building docker image ${IMAGE_NAME}:${TAG}"
	docker build -t ${IMAGE_NAME}:${TAG} .

minikube-load:
	@echo "--------------------------------------Load image to minikube: ${IMAGE_NAME}:${TAG}"
	minikube kubectl -- delete -f ${YAML}
	minikube image load ${IMAGE_NAME}:${TAG} --overwrite=true
	minikube kubectl -- apply -f ${YAML}

all: image-runnable #docker-build minikube-load

clean:
	rm -rf build/${RUN_NAME}