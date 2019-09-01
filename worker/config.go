package worker

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ConsulConfig struct {
	Path string `json:"consulPath"`
}

var (
	ConsulConf *ConsulConfig
)

//初始化配置文件
func InitConfig(fileName string) (err error) {
	execpath, err := os.Executable() // 获得程序路径
	if err != nil {
		panic("获取程序路径出错" + err.Error())
	}
	configfile := filepath.Join(filepath.Dir(execpath), "./"+fileName)
	content, err := ioutil.ReadFile(configfile)
	if err != nil {
		return
	}
	err = json.Unmarshal(content, &ConsulConf)
	if err != nil {
		return
	}
	return
}
