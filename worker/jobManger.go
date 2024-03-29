package worker

import (
	"go-cron/common"

	consulwatch "github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	g_JobManger *JobManger
)

type JobManger struct {
	Client  *consulapi.Client
	KV      *consulapi.KV
	Session *consulapi.Session
	Agent   *consulapi.Agent
}

func InitJobManger() (err error) {
	config := &consulapi.Config{
		Address: consulConf.Path,
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		common.Logger.Error("连接consul失败", zap.Error(err))
		return
	}
	g_JobManger = &JobManger{
		Client:  client,
		KV:      client.KV(),
		Session: client.Session(),
		Agent:   client.Agent(),
	}
	go g_JobManger.watchCronJob()
	return
}

//监控consul 任务变化
func (jobManger *JobManger) watchCronJob() (err error) {
	params := make(map[string]interface{})
	params["type"] = "keyprefix"
	params["prefix"] = common.JOB_SAVE_DIR
	plan, err := consulwatch.Parse(params)
	plan.Handler = g_JobManger.buildChangeEvent
	if err != nil {
		common.Logger.Error("watchPlan转换失败", zap.Error(err))
	}
	err = plan.Run(consulConf.Path)
	if err != nil {
		common.Logger.Error("watch失败", zap.Error(err))
	}
	return
}

func (jobManger *JobManger) buildChangeEvent(idx uint64, result interface{}) {
	kvpairs, ok := result.(consulapi.KVPairs)
	if ok {
		g_JobScheduler.newKVPairs = kvpairs
		g_JobScheduler.pushChangeEvent()
	}
}

// GetLock 获取锁
func (jobManger *JobManger) GetLock(lockName string) (bool, error) {
	//分布式锁
	sessionOpts := &consulapi.SessionEntry{
		Behavior: consulapi.SessionBehaviorDelete,
		TTL:      "10s",
	}
	session, _, err := jobManger.Session.Create(sessionOpts, nil)
	if err != nil {
		return false, err
	}
	kvPair := &consulapi.KVPair{
		Key:     common.JOB_LOCK_DIR + lockName,
		Session: session,
	}
	l, _, err := jobManger.KV.Acquire(kvPair, nil)
	if err != nil {
		return false, err
	}
	// opts := &consulapi.LockOptions{
	// 	Key:          common.JOB_LOCK_DIR + lockName,
	// 	SessionTTL:   "10s",
	// 	LockTryOnce:  true,
	// 	LockWaitTime: 1 * time.Millisecond,
	// 	SessionOpts:  sessionOpts,
	// }
	// l, err = jobManger.Client.LockOpts(opts)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	return l, err
}
