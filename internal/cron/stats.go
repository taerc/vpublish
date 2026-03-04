package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/taerc/vpublish/internal/repository"
)

type StatsJob struct {
	statsRepo *repository.StatsRepository
}

func NewStatsJob(statsRepo *repository.StatsRepository) *StatsJob {
	return &StatsJob{statsRepo: statsRepo}
}

// AggregateDaily 每日聚合任务
// 将下载日志聚合到统计表
func (j *StatsJob) AggregateDaily() {
	// TODO: 实现从 download_logs 聚合到 download_stats 的逻辑
	log.Println("Running daily stats aggregation...")
}

// StartCronJobs 启动定时任务
func StartCronJobs(statsRepo *repository.StatsRepository) *cron.Cron {
	c := cron.New(cron.WithLocation(time.Local))
	statsJob := NewStatsJob(statsRepo)

	// 每天凌晨1点执行聚合任务
	c.AddFunc("0 1 * * *", statsJob.AggregateDaily)

	c.Start()
	return c
}
