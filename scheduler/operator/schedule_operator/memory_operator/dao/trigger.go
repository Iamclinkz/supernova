package dao

import (
	"github.com/google/btree"
	"supernova/scheduler/model"
)

type Trigger struct {
	model.Trigger
}

func FromModelTrigger(mTrigger *model.Trigger) *Trigger {
	return &Trigger{
		Trigger: *mTrigger,
	}
}

func (t *Trigger) ToModelOnFireLog() *model.Trigger {
	ret := new(model.Trigger)
	*ret = t.Trigger
	return ret
}

func (t *Trigger) Less(than btree.Item) bool {
	return t.TriggerNextTime.Before(than.(*Trigger).TriggerNextTime)
}
