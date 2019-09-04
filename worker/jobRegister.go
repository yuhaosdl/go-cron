package worker

import (
	"fmt"
	"go-cron/common"
	"net"
	"time"

	"go.uber.org/zap"

	consulapi "github.com/hashicorp/consul/api"
)

//InitRegister 初始化服务注册
func InitRegister() (err error) {
	check := &consulapi.AgentServiceCheck{
		TTL: fmt.Sprintf("%ds", consulConf.TTL),
		DeregisterCriticalServiceAfter: "1m",
	}
	service := &consulapi.AgentServiceRegistration{
		Name:  consulConf.Name,
		ID:    fmt.Sprintf("%s(%s)", consulConf.Name, GetLocalAddress()),
		Check: check,
	}
	err = g_JobManger.Agent.ServiceRegister(service)
	if err != nil {
		common.Logger.Error("服务注册失败", zap.Error(err))
		return
	}
	//注册worker服务 心跳
	go func(checkID string) {
		keepAliveTicker := time.NewTicker(time.Duration(consulConf.TTL) * time.Second / 5)
		for {
			<-keepAliveTicker.C
			err := g_JobManger.Agent.PassTTL(checkID, "")
			if err != nil {
				common.Logger.Error("续租失败", zap.Error(err))
			}
		}
	}("service:" + service.ID)
	return
}

//GetLocalAddress 获取ip地址
func GetLocalAddress() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		common.Logger.Error("net.Interfaces failed", zap.Error(err))
	}
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}
