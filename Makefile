all:
	cd scheduler && make all
	cd executor-example/http-executor && make all
	cd executor-example/shell-executor && make all
	@echo "finished make all"

k8s:
	cd executor-example/http-executor-k8s && make k8s
	cd scheduler && make k8s
	@echo "finished make k8s"
	sleep 40
	zsh -c "kubectl port-forward -n supernova svc/scheduler-service 8080:8080 &"

test: all
	cd tests/functional-test && go test stress_test.go
	cd tests/functional-test && go test fail_test.go
	cd tests/functional-test && go test kill_scheduler_test.go
	cd tests/functional-test && go test executor_graceful_stop_test.go
	@echo "test finished"

k8s-test: k8s
	cd stress-test/k8s-tests && go test stress_test.go
