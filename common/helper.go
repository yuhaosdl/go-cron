package common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job, modifyIndex uint64) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	// 解析JOB的cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	// 生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:         job,
		Expr:        expr,
		NextTime:    expr.Next(time.Now()),
		ModifyIndex: modifyIndex,
	}
	return
}

// 构造执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计算调度时间
		RealTime: time.Now(),               // 真实调度时间
	}
	if jobSchedulePlan.Job.Type == "shell" {
		jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	}
	return
}
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

func CovertToJobSchedulePlan(value []byte, modifyIndex uint64) (jobSchedulePlan *JobSchedulePlan, err error) {
	job := &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	return BuildJobSchedulePlan(job, modifyIndex)
}
