RUN_NAME := http-executor

build-binary:
	@echo "--------------------------------------Building ${RUN_NAME} binary..."
	go build -o build/bin/${RUN_NAME} *.go

all: build-binary

clean:
	rm -rf build/${RUN_NAME}