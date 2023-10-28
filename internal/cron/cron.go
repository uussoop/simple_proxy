package cron

import (
	"github.com/robfig/cron"
	"github.com/rodrikv/openai_proxy/database"
)

var CronJob *cron.Cron

func init() {
	CronJob = cron.New()
	initJobs()
}

func initJobs() {
	if CronJob != nil {
		CronJob.AddFunc("0 0 12 * * *", database.ResetUsageToday)
		CronJob.AddFunc("0 * * * *", database.ResetRequestCount)
		CronJob.AddFunc("0 0 12 * *", database.ResetEndpointDailyUsage)
		CronJob.AddFunc("0 * * * *", database.ResetEndpointUsage)
	}
}

func Start() {
	if CronJob != nil {
		CronJob.Start()
	}
}
