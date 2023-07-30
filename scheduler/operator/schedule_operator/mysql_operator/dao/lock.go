package dao

type Lock struct {
	LockName string `gorm:"column:lock_name;type:varchar(64);primarykey"`
}

func (j *Lock) TableName() string {
	return "t_lock"
}
