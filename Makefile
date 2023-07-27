all:
	cd scheduler && make all
	cd executor-example/http-executor && make all
	cd executor-example/shell-executor && make all
	@echo "finished"