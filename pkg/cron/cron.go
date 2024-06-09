package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/pkg/cron/jobs"
)

var CronJob *cron.Cron

func init() {

	CronJob = cron.New(cron.WithSeconds())

	initJobs()
}

func initJobs() {
	if CronJob != nil {

		// 0 0 0 ? * * * everyday
		// 0 0/1 * * * ? every minute
		CronJob.AddFunc("* * * ? * * *", jobs.SaveUsage)
		CronJob.AddFunc("0 0 0 ? * * *", database.ResetUsageToday)

	}
}

func Start() {
	if CronJob != nil {
		CronJob.Start()
	}
}
