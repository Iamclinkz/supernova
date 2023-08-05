package simple_http_server

const mainViewTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Simple HTTP Server</title>
	<style>
		table {
			border-collapse: collapse;
			width: 100%;
		}
		th, td {
			border: 1px solid black;
			padding: 8px;
			text-align: left;
		}
		th {
			background-color: #f2f2f2;
		}
	</style>
	<script>
		function shutdownServer() {
			fetch('/shutdown').then(function(response) {
				if (response.ok) {
					alert('Server is shutting down...');
				} else {
					alert('Error shutting down server');
				}
			});
		}
	</script>
</head>
<body>
<h1>测试配置</h1>
<table>
	<tr>
		<th>字段</th>
		<th>值</th>
	</tr>
	<tr>
		<td>随机失败概率</td>
		<td>{{printf "%.2f" .ServeConfig.FailRate}}</td>
	</tr>
	<tr>
		<td>监听端口</td>
		<td>{{.ServeConfig.ListeningPort}}</td>
	</tr>
	<tr>
		<td>测试总Trigger数量</td>
		<td>{{.ServeConfig.TriggerCount}}</td>
	</tr>
	<tr>
		<td>允许任务成功后仍然请求</td>
		<td>{{.ServeConfig.AllowDuplicateCalled}}</td>
	</tr>
	<tr>
		<td>失败一次后必定成功</td>
		<td>{{.ServeConfig.SuccessAfterFirstFail}}</td>
	</tr>
</table>
<h1>任务执行情况</h1>
<table>
	<tr>
		<th>字段</th>
		<th>值</th>
	</tr>
	<tr>
		<td>返回成功数量</td>
		<td>{{.Result.SuccessCount}}</td>
	</tr>
	<tr>
		<td>未成功TriggerID</td>
		<td>{{.Result.HaveNotCalledCount}}</td>
	</tr>
	<tr>
		<td>未成功数量</td>
		<td>{{.Result.CalledButFailCount}}</td>
	</tr>
	<tr>
		<td>收到请求次数</td>
		<td>{{.Result.CalledTotal}}</td>
	</tr>
	<tr>
		<td>一次都没请求的TriggerID</td>
		<td>{{.Result.UncalledTriggers}}</td>
	</tr>
	<tr>
		<td>仍没有成功的TriggerID</td>
		<td>{{.Result.FailedTriggers}}</td>
	</tr>
	<tr>
		<td>失败率</td>
		<td>{{.Result.FailTriggerRate}}</td>
	</tr>
	<tr>
		<td>请求多次的TriggerID</td>
		<td>{{.Result.CalledTwiceOrMore}}</td>
	</tr>
	<tr>
		<td>第一条请求到达时间</td>
		<td>{{.Result.FirstRequestTime.Format "2006-01-02 15:04:05"}}</td>
	</tr>
	<tr>
		<td>最后一条请求到达时间</td>
		<td>{{.Result.LastRequestTime.Format "2006-01-02 15:04:05"}}</td>
	</tr>
	<tr>
		<td>平均每秒多少条请求</td>
		<td>{{printf "%.1f" .Result.AvgRequestsPerSecond}}</td>
	</tr>
</table>
<button onclick="window.open('/executor-log', '_blank')">查看优雅退出Executor日志</button>
<button onclick="window.open('/scheduler-log', '_blank')">查看被杀死的Scheduler日志</button>
<button onclick="shutdownServer()">关闭服务器</button>
</body>
</html>
`
