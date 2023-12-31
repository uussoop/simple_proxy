package cron

import (
	"github.com/robfig/cron"
	"github.com/uussoop/simple_proxy/pkg/cron/jobs"
)

var CronJob *cron.Cron

func init() {
	CronJob = cron.New()

	initJobs()
}

func initJobs() {
	if CronJob != nil {
		// every 5 minutes get models
		CronJob.AddFunc("*/60 * * * *", jobs.SaveUsage)

	}
}

func Start() {
	if CronJob != nil {
		CronJob.Start()
	}
}
