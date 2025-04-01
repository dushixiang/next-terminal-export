package model

import (
	"github.com/dushixiang/next-terminal-export/common"
)

type Tenant struct {
	ID      string          `gorm:"primary_key,type:varchar(36)" json:"id"`
	Name    string          `gorm:"type:varchar(500)" json:"name"`
	Created common.JsonTime `json:"created"`
}

func (r *Tenant) TableName() string {
	return "tenants"
}
