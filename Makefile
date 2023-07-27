all:
	cd scheduler && make all
	cd executor-example/http-executor && make all
	cd executor-example/shell-executor && make all
	@echo "finished make all"

test: all
	cd stress-test/tests && go test stress_test.go
	cd stress-test/tests && go test fail_test.go
	cd stress-test/tests && go test kill_scheduler_test.go
	cd stress-test/tests && go test executor_graceful_stop_test.go
	@echo "test finished"
