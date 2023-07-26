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

