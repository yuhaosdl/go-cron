package worker

import (
	"fmt"
	"go-cron/common"
	"math/rand"
	"os/exec"
	"time"

	"github.com/valyala/fasthttp"
)

// 任务执行器
type Executor struct {
}

var (
	g_executor *Executor
)

// ExecuteJob 执行任务
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			err    error
			output []byte
		)
		// 任务结果
		result := &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}

		// 记录任务开始时间
		result.StartTime = time.Now()

		// 上锁
		// 随机睡眠(0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		// 初始化分布式锁
		l, err := g_JobManger.GetLock(info.Job.Name + info.PlanTime.String())
		if err != nil {
			fmt.Println(err)
		}
		// stopCh := make(chan struct{})
		// leaderCh, err := l.Lock(stopCh)
		// defer func(l *consulapi.Lock) {
		// 	close(l.sessionRenew)
		// 	l.sessionRenew = nil
		// }(l)
		//l.Unlock()

		if !l || err != nil { // 上锁失败
			fmt.Println("获取锁失败了")
			fmt.Println(err)
			result.Err = err
			result.EndTime = time.Now()
		} else {
			fmt.Println("获取了锁")
			// 上锁成功后，重置任务启动时间
			result.StartTime = time.Now()
			if info.Job.Type == "shell" {
				output, err = executor.ExecuteShellJob(info)
			} else {
				output, err = executor.ExecuteHttpJob(info)
			}
			// 记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// 任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingTable中删除掉执行记录
		g_JobScheduler.pushJobResult(result)
	}()
}

// ExecuteShellJob 执行shell命令
func (executor *Executor) ExecuteShellJob(info *common.JobExecuteInfo) (output []byte, err error) {
	// 执行shell命令
	cmd := exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

	// 执行并捕获输出
	output, err = cmd.CombinedOutput()

	return
}

// ExecuteHttpJob 执行Http任务
func (executor *Executor) ExecuteHttpJob(info *common.JobExecuteInfo) (output []byte, err error) {

	url := info.Job.Url

	// 填充表单，类似于net/url
	args := &fasthttp.Args{}

	status, resp, err := fasthttp.Post(nil, url, args)
	if err != nil {
		fmt.Println("请求失败:", err.Error())
		return
	}

	if status != fasthttp.StatusOK {
		fmt.Println("请求没有成功:", status)
		return
	}
	output = resp
	fmt.Println(string(resp))
	return
}

// InitExecutor 初始化执行器
func InitExecutor() (err error) {
	g_executor = &Executor{}
	return
}
