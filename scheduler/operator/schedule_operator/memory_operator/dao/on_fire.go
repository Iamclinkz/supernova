package dao

import (
	"github.com/google/btree"
	"supernova/scheduler/model"
)

type OnFireLog struct {
	model.OnFireLog
}

func FromModelOnFireLog(mOnFireLog *model.OnFireLog) *OnFireLog {
	return &OnFireLog{
		OnFireLog: *mOnFireLog,
	}
}

func (o *OnFireLog) ToModelOnFireLog() *model.OnFireLog {
	ret := new(model.OnFireLog)
	*ret = o.OnFireLog
	return ret
}

func (o *OnFireLog) Less(than btree.Item) bool {
	return o.RedoAt.Before(than.(*OnFireLog).RedoAt)
}
