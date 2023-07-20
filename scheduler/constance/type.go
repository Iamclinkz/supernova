package constance

import "strconv"

type ScheduleType int8

const (
	ScheduleTypeMin  ScheduleType = iota
	ScheduleTypeNone              //一次都不执行（等待用户手动执行）
	ScheduleTypeOnce              //执行一次
	ScheduleTypeCron              //cron表达式
	ScheduleTypeMax
)

func (t ScheduleType) String() string {
	switch t {
	case ScheduleTypeNone:
		return "ScheduleTypeNone"
	case ScheduleTypeOnce:
		return "ScheduleTypeOnce"
	case ScheduleTypeCron:
		return "ScheduleTypeCron"
	default:
		return "UnknownScheduleType" + strconv.Itoa(int(t))
	}
}

func (t ScheduleType) Valid() bool {
	return t > ScheduleTypeMin && t < ScheduleTypeMax
}

type MisfireStrategyType int8

const (
	MisfireStrategyTypeMin        MisfireStrategyType = iota
	MisfireStrategyTypeDoNothing                      //错过了本次就不fire了
	MisfireStrategyTypeFireNow                        //加急，直接fire
	MisfireStrategyTypeFireNormal                     //正常fire，扔到定时器中，下次tick正常fire
	MisfireStrategyTypeMax
)

func (t MisfireStrategyType) String() string {
	switch t {
	case MisfireStrategyTypeDoNothing:
		return "MisfireStrategyTypeDoNothing"
	case MisfireStrategyTypeFireNow:
		return "MisfireStrategyTypeFireNow"
	case MisfireStrategyTypeFireNormal:
		return "MisfireStrategyTypeFireNormal"
	default:
		return "UnknownMisfireStrategyType" + strconv.Itoa(int(t))
	}
}

func (t MisfireStrategyType) Valid() bool {
	return t > MisfireStrategyTypeMin && t < MisfireStrategyTypeMax
}

type ExecutorRouteStrategyType int8

const (
	ExecutorRouteStrategyTypeMin       ExecutorRouteStrategyType = iota
	ExecutorRouteStrategyTypeRandom                              //随机选择一个执行器
	ExecutorRouteStrategyTypeHash                                //一致性路由，本trigger最好分到同一个执行器实例执行
	ExecutorRouteStrategyTypeBroadcast                           //分给全部执行器执行
	ExecutorRouteStrategyTypeMax
)

func (t ExecutorRouteStrategyType) String() string {
	switch t {
	case ExecutorRouteStrategyTypeRandom:
		return "ExecutorRouteStrategyTypeRandom"
	case ExecutorRouteStrategyTypeHash:
		return "ExecutorRouteStrategyTypeHash"
	case ExecutorRouteStrategyTypeBroadcast:
		return "ExecutorRouteStrategyTypeBroadcast"
	default:
		return "UnknownExecutorRouteStrategyType" + strconv.Itoa(int(t))
	}
}

func (t ExecutorRouteStrategyType) Valid() bool {
	return t > ExecutorRouteStrategyTypeMin && t < ExecutorRouteStrategyTypeMax
}

type JobStatus int8

const (
	JobStatusMin     JobStatus = iota
	JobStatusNormal            //正常等待执行
	JobStatusStopped           //用户手动停止。这种情况下只有失败的任务还可以重试，正常的trigger暂停执行
	JobStatusDeleted           //用户手动删掉了任务，但是选择了正在执行的重试trigger不受影响
	JobStatusMax
)

func (t JobStatus) String() string {
	switch t {
	case JobStatusStopped:
		return "JobStatusStopped"
	case JobStatusNormal:
		return "JobStatusNormal"
	case JobStatusDeleted:
		return "JobStatusAcquired"
	default:
		return "UnknownJobStatus" + strconv.Itoa(int(t))
	}
}

func (t JobStatus) Valid() bool {
	return t > JobStatusMin && t < JobStatusMax
}

type OnFireStatus int8

const (
	OnFireStatusMin       OnFireStatus = iota
	OnFireStatusWaiting                //已经被某个Scheduler取走，正在等待分配执行器
	OnFireStatusExecuting              //正在被某个Executor执行
	OnFireStatusFinished               //执行结束（可能是执行成功，也可能是执行失败，且重试次数用光）
	OnFireStatusMax
)

func (t OnFireStatus) String() string {
	switch t {
	case OnFireStatusWaiting:
		return "OnFireStatusWaiting"
	case OnFireStatusExecuting:
		return "OnFireStatusExecuting"
	case OnFireStatusFinished:
		return "OnFireStatusFinished"
	default:
		return "UnknownOnFireStatus" + strconv.Itoa(int(t))
	}
}

func (t OnFireStatus) Valid() bool {
	return t > OnFireStatusMin && t < OnFireStatusMax
}

type TriggerStatus int8

const (
	TriggerStatusMin    TriggerStatus = iota
	TriggerStatusNormal               //正常状态
	TriggerStatusError                //出错（例如用户指定了错误的Cron表达式）
	TriggerStatusPause                //用户手动停止
	TriggerStatusMax
)

func (t TriggerStatus) String() string {
	switch t {
	case TriggerStatusNormal:
		return "TriggerStatusNormal"
	case TriggerStatusError:
		return "TriggerStatusError"
	case TriggerStatusPause:
		return "TriggerStatusPause"
	default:
		return "UnknownTriggerStatus" + strconv.Itoa(int(t))
	}
}

func (t TriggerStatus) Valid() bool {
	return t > TriggerStatusMin && t < TriggerStatusMax
}
