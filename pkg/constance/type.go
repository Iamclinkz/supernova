package constance

import "strconv"

type ScheduleType int8

const (
	ScheduleTypeMin ScheduleType = iota
	ScheduleTypeNone
	ScheduleTypeOnce
	ScheduleTypeCron
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
	MisfireStrategyTypeMin MisfireStrategyType = iota
	MisfireStrategyTypeDoNothing
	MisfireStrategyTypeFireNow
	MisfireStrategyTypeFireNormal
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
	ExecutorRouteStrategyTypeMin ExecutorRouteStrategyType = iota
	ExecutorRouteStrategyTypeRandom
	ExecutorRouteStrategyTypeTag
	ExecutorRouteStrategyTypeBroadcast
	ExecutorRouteStrategyTypeMax
)

func (t ExecutorRouteStrategyType) String() string {
	switch t {
	case ExecutorRouteStrategyTypeRandom:
		return "ExecutorRouteStrategyTypeRandom"
	case ExecutorRouteStrategyTypeTag:
		return "ExecutorRouteStrategyTypeTag"
	case ExecutorRouteStrategyTypeBroadcast:
		return "ExecutorRouteStrategyTypeBroadcast"
	default:
		return "UnknownExecutorRouteStrategyType" + strconv.Itoa(int(t))
	}
}

func (t ExecutorRouteStrategyType) Valid() bool {
	return t > ExecutorRouteStrategyTypeMin && t < ExecutorRouteStrategyTypeMax
}

type GlueType int8

const (
	GlueTypeMin GlueType = iota
	GlueTypeShell
	GlueTypePython
	GlueTypeGo
	GlueTypeMax
)

func (t GlueType) String() string {
	switch t {
	case GlueTypeShell:
		return "GlueTypeShell"
	case GlueTypePython:
		return "GlueTypePython"
	case GlueTypeGo:
		return "GlueTypeGo"
	default:
		return "UnknownGlueType" + strconv.Itoa(int(t))
	}
}

func (t GlueType) Valid() bool {
	return t > GlueTypeMin && t < GlueTypeMax
}

type ExecutorBlockStrategy int8

const (
	ExecutorBlockStrategyMin ExecutorBlockStrategy = iota
	ExecutorBlockStrategyBlock
	ExecutorBlockStrategyNoBlock
	ExecutorBlockStrategyMax
)

func (t ExecutorBlockStrategy) String() string {
	switch t {
	case ExecutorBlockStrategyBlock:
		return "ExecutorBlockStrategyBlock"
	case ExecutorBlockStrategyNoBlock:
		return "ExecutorBlockStrategyNoBlock"
	default:
		return "UnknownExecutorBlockStrategy" + strconv.Itoa(int(t))
	}
}

func (t ExecutorBlockStrategy) Valid() bool {
	return t > ExecutorBlockStrategyMin && t < ExecutorBlockStrategyMax
}

type JobStatus int8

const (
	JobStatusMin JobStatus = iota
	JobStatusStopped
	JobStatusWaiting
	JobStatusAcquired
	JobStatusMax
)

func (t JobStatus) String() string {
	switch t {
	case JobStatusStopped:
		return "JobStatusStopped"
	case JobStatusWaiting:
		return "JobStatusWaiting"
	case JobStatusAcquired:
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
	OnFireStatusMin OnFireStatus = iota
	OnFireStatusWaiting
	OnFireStatusExecuting
	OnFireStatusFinished
	OnFireStatusFailed
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
	case OnFireStatusFailed:
		return "OnFireStatusFailed"
	default:
		return "UnknownOnFireStatus" + strconv.Itoa(int(t))
	}
}

func (t OnFireStatus) Valid() bool {
	return t > OnFireStatusMin && t < OnFireStatusMax
}

type TriggerStatus int8

const (
	TriggerStatusMin TriggerStatus = iota
	TriggerStatusNormal
	TriggerStatusError
	TriggerStatusPause
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
