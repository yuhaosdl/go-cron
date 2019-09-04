package worker

import (
	"fmt"
	"go-cron/common"
	"time"

	"go.uber.org/zap"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	g_JobScheduler *Scheduler
)

type Scheduler struct {
	changeChan chan string
	//killChan          chan *common
	jobResultChan     chan *common.JobExecuteResult
	planTable         map[string]*common.JobSchedulePlan
	newKVPairs        consulapi.KVPairs
	jobExecutingTable map[string]*common.JobExecuteInfo
}

func InitScheduler() (err error) {
	g_JobScheduler = &Scheduler{
		changeChan:        make(chan string, 1000),
		planTable:         make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}
	// 启动调度协程
	go g_JobScheduler.scheduleLoop()
	return
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop() {

	// 初始化一次(1秒)
	scheduleAfter := scheduler.schedule()

	// 调度的延迟定时器
	scheduleTimer := time.NewTimer(scheduleAfter)

	// 定时任务common.Job
	for {
		select {
		case <-scheduler.changeChan: //监听任务变化事件
			// 对内存中维护的任务列表做变更
			scheduler.handleChangeEvent()
		case jobResult := <-scheduler.jobResultChan: // 监听任务执行结果
			scheduler.handleJobResult(jobResult)
		case <-scheduleTimer.C: // 最近的任务到期了

		}
		// 调度一次任务
		scheduleAfter = scheduler.schedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	// 删除执行状态
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
	//日志记录结果
	if result.Err != nil {
		common.Logger.Error(fmt.Sprintf("执行任务（%s）失败", result.ExecuteInfo.Job.Name), zap.Error(result.Err))
	} else {
		common.Logger.Info(fmt.Sprintf("执行任务（%s）成功", result.ExecuteInfo.Job.Name), zap.String("output", result.Output), zap.Time("startTime", result.StartTime), zap.Time("endTime", result.EndTime))
	}

}

// 重新计算任务调度状态
func (scheduler *Scheduler) schedule() (scheduleAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)

	// 如果任务表为空话，随便睡眠多久
	if len(scheduler.planTable) == 0 {
		common.Logger.Info("没有任务")
		scheduleAfter = 1 * time.Second
		return
	}

	// 当前时间
	now = time.Now()

	// 遍历所有任务
	for _, jobPlan = range scheduler.planTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now) // 更新下次执行时间
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)
	return
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {

	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

	// 如果任务正在执行，跳过本次调度
	if _, jobExecuting := scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		common.Logger.Info(fmt.Sprintf("尚未退出,跳过执行:%s", jobPlan.Job.Name))
		return
	}

	// 构建执行状态信息
	jobExecuteInfo := common.BuildJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	common.Logger.Info(fmt.Sprintf("执行任务:%s", jobExecuteInfo.Job.Name), zap.Time("PlanTime", jobExecuteInfo.PlanTime), zap.Time("RealTime", jobExecuteInfo.RealTime))
	g_executor.ExecuteJob(jobExecuteInfo)
}

// 处理change事件
func (scheduler *Scheduler) handleChangeEvent() {
	newPlanTable := make(map[string]*common.JobSchedulePlan)
	for _, kvpair := range scheduler.newKVPairs {
		if kvpair.Key == common.JOB_SAVE_DIR {
			continue
		}
		jobName := common.ExtractJobName(kvpair.Key)
		if plan, exists := scheduler.planTable[jobName]; exists {
			if plan.ModifyIndex == kvpair.ModifyIndex {
				newPlanTable[jobName] = plan
			} else {
				if newplan, err := common.CovertToJobSchedulePlan(kvpair.Value, kvpair.ModifyIndex); err == nil {
					newPlanTable[jobName] = newplan
				}
			}

		} else {
			if newplan, err := common.CovertToJobSchedulePlan(kvpair.Value, kvpair.ModifyIndex); err == nil {
				newPlanTable[jobName] = newplan
			}
		}
	}
	scheduler.planTable = newPlanTable
}

//处理 强杀事件
func (scheduler *Scheduler) handleKillEvent() {
	// if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
	// 	jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程, 任务得到退出
	// }
}

//触发change事件
func (scheduler *Scheduler) pushChangeEvent() {
	scheduler.changeChan <- "1"
}
func (scheduler *Scheduler) pushJobResult(result *common.JobExecuteResult) {
	scheduler.jobResultChan <- result
}
