package worker

import (
	"fmt"
	"gostudy/go-cron/common"
	"time"

	consulwatch "github.com/hashicorp/consul/api/watch"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	g_JobManger *JobManger
)

type JobManger struct {
	Client  *consulapi.Client
	KV      *consulapi.KV
	Session *consulapi.Session
}

func InitJobManger() (err error) {
	config := &consulapi.Config{
		Address: ConsulConf.Path,
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Println("连接consul失败")
		return
	}
	g_JobManger = &JobManger{
		Client:  client,
		KV:      client.KV(),
		Session: client.Session(),
	}
	go g_JobManger.watchCronJob()
	return
}

//监控consul 任务变化
func (jobManger *JobManger) watchCronJob() (err error) {
	params := make(map[string]interface{})
	params["type"] = "keyprefix"
	params["prefix"] = "cron/jobs/"
	plan, err := consulwatch.Parse(params)
	plan.Handler = g_JobManger.buildChangeEvent
	if err != nil {
		fmt.Println(err)
	}
	plan.Run(ConsulConf.Path)
	return
}

func (jobManger *JobManger) buildChangeEvent(idx uint64, result interface{}) {
	kvpairs, ok := result.(consulapi.KVPairs)
	if ok {
		g_JobScheduler.newKVPairs = kvpairs
		for _, v := range kvpairs {
			fmt.Printf("%+v\n", *v)
		}
		g_JobScheduler.pushChangeEvent()
	}
}

func (jobManger *JobManger) GetLock(jobName string) (l *consulapi.Lock, err error) {
	opts := &consulapi.LockOptions{
		Key:          common.JOB_LOCK_DIR + jobName,
		SessionTTL:   "10s",
		LockTryOnce:  true,
		LockWaitTime: 1 * time.Millisecond,
	}
	l, err = jobManger.Client.LockOpts(opts)
	if err != nil {
		fmt.Println(err)
	}
	return
}
