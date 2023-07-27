# 测试说明

## 1.压力测试

```sh
 go test -v stress_test.go
```

## 2.极限情况失败测试

```sh
go test -v fail_test.go
```

## 3.Scheduler宕机测试

```sh
go test -v kill_scheduler_test.go
```

## 4.Executor优雅退出测试

```sh
go test -v executor_graceful_stop_test.go
```

## 1000.加了优雅测试之后强制杀死测试进程。。因为ctrl+c已经不好使了。。

```sh
ps -ef | grep 'go test -v' | grep -v grep | awk '{print $2}' | xargs kill
```