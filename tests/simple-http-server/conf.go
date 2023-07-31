package simple_http_server

//监听/test的POST方法，记录并统计收到的http请求中的triggerID字段的值，根据fail字段随机失败。
//如果失败，返回500，Error，如果不失败，返回200，OK。无论失败与否都会记录数据。

type SimpleHttpServerInitConf struct {
	FailRate              float32 //随机失败率是多少。如果是0，则不失败
	ListeningPort         int     //监听的端口
	TriggerCount          int     //应该来的trigger
	AllowDuplicateCalled  bool    //是否允许一个trigger执行多次？（需要结合 “最多一次“ 和 ”最少一次“ 语义指定）
	SuccessAfterFirstFail bool    //第一次失败之后，之后是否必须成功
}

type SimpleHttpServerCheckConf struct {
	AllSuccess                    bool    //一定要都成功
	NoUncalledTriggers            bool    //一定不能有没有执行过的trigger
	FailTriggerRateNotGreaterThan float32 //失败率不能高于
}
