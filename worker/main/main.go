package main

import (
	"go-cron/common"
	cron_work "go-cron/worker"
	"time"

	"go.uber.org/zap"
)

func main() {
	err := cron_work.InitConfig("work.json")
	if err != nil {
		common.Logger.Error("初始化配置出错", zap.Error(err))
		return
	}
	err = cron_work.InitScheduler()
	if err != nil {
		common.Logger.Error("初始化JobScheduler失败", zap.Error(err))
		return
	}
	err = cron_work.InitExecutor()
	if err != nil {
		common.Logger.Error("初始化Executor失败", zap.Error(err))
		return
	}
	err = cron_work.InitJobManger()
	if err != nil {
		common.Logger.Error("初始化Consul失败", zap.Error(err))
		return
	}
	err = cron_work.InitRegister()
	if err != nil {
		common.Logger.Error("注册worker失败", zap.Error(err))
		return
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
