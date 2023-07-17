package model

import "time"

type UserDetail struct {
	ID        int64     `json:"-"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"-"`
	Updated   int64     `json:"-"`
}

func (d *UserDetail) NeedUpdate() bool {
	return d.Name != "" || d.Phone != ""
}
