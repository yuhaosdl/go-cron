package common

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	JOB_EVENT_CHANGE = 1
	JOB_EVENT_KILL   = 2
	JOB_SAVE_DIR     = "cron/jobs/"
	JOB_Kill_DIR     = "cron/kill/"
	JOB_LOCK_DIR     = "cron/lock/"
)

type Config struct {
	ConsulPath []string
}

// 定时任务
type Job struct {
	Name     string `json:"name"`     //  任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
	Type     string `json:"type"`     //命令类型  http请求/shell命令
	Url      string `json:"url"`      //当类型为http时执行的命令
}

// 任务调度计划
type JobSchedulePlan struct {
	Job         *Job                 // 要调度的任务信息
	Expr        *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime    time.Time            // 下次调度时间
	ModifyIndex uint64               //变更版本
}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context
	CancelFunc context.CancelFunc //  用于取消command执行的cancel函数
}
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	Output      []byte          // 输出
	Err         error           // 脚本错误原因
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       // 结束时间
}
