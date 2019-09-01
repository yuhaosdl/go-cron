package main

import (
	"fmt"
	cron_work "gostudy/go-cron/worker"
	"time"
)

func main() {
	err := cron_work.InitConfig("work.json")
	if err != nil {
		fmt.Println("初始化配置出错")
		return
	}
	err = cron_work.InitScheduler()
	if err != nil {
		fmt.Println("初始化JobScheduler失败")
		return
	}
	err = cron_work.InitExecutor()
	if err != nil {
		fmt.Println("初始化Executor失败")
		return
	}
	err = cron_work.InitJobManger()
	if err != nil {
		fmt.Println("初始化Consul失败")
		return
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
