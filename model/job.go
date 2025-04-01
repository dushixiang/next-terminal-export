package model

import (
	"github.com/dushixiang/next-terminal-export/common"
)

type Job struct {
	ID          string          `gorm:"primary_key,type:varchar(36)" json:"id"`
	CronJobId   int             `json:"cronJobId"`
	Name        string          `gorm:"type:varchar(500)" json:"name"`
	Func        string          `gorm:"type:varchar(200)" json:"func"`
	Cron        string          `gorm:"type:varchar(100)" json:"cron"`
	Mode        string          `gorm:"type:varchar(50)" json:"mode"`
	ResourceIds string          `json:"resourceIds"`
	Status      string          `gorm:"type:varchar(20)" json:"status"`
	Metadata    string          `json:"metadata"`
	Created     common.JsonTime `json:"created"`
	Updated     common.JsonTime `json:"updated"`
}

func (r *Job) TableName() string {
	return "jobs"
}

type JobLog struct {
	ID        string          `json:"id"`
	Timestamp common.JsonTime `json:"timestamp"`
	JobId     string          `json:"jobId"`
	Message   string          `json:"message"`
}

func (r *JobLog) TableName() string {
	return "job_logs"
}
