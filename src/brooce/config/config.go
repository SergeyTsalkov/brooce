package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"brooce/myip"
)

type ConfigType struct {
	ClusterName string `json:"cluster_name"`
	ProcName    string `json:"process_name"`

	Web struct {
		Addr     string
		Username string
		Password string
	}

	Redis struct {
		Host     string
		Password string
	}
	Suicide struct {
		Enabled bool
		Command string
		Time    int
	}
	Syslog struct {
		Host string
	}
	Queues map[string]int
	Path   string
}

var Config ConfigType

func (c *ConfigType) TotalThreads() (threads int) {
	for _, ct := range Config.Queues {
		threads += ct
	}
	return
}

func init() {
	bytes, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".brooce"))
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(bytes, &Config)
	if err != nil {
		log.Fatalln(err)
	}

	if Config.ClusterName == "" {
		Config.ClusterName = "brooce"
	}

	if Config.ProcName == "" {
		Config.ProcName = fmt.Sprintf("%v-%v", myip.PublicIPv4(), os.Getpid())
	}

	if Config.Web.Addr == "" {
		Config.Web.Addr = ":8080"
	}

	if Config.Queues == nil {
		log.Fatalln("The queues hash was not configured in the ~/.brooce config file!")
	}

	init_redis()
	init_path()
	init_suicide()
}

func init_redis() {
	if Config.Redis.Host == "" {
		Config.Redis.Host = "localhost"
	}

	if !strings.Contains(Config.Redis.Host, ":") {
		Config.Redis.Host = Config.Redis.Host + ":6379"
	}
}

func init_path() {
	if Config.Path == "" {
		return
	}

	extrapath := Config.Path
	if !strings.HasPrefix(extrapath, "/") {
		extrapath = filepath.Join(os.Getenv("HOME"), extrapath)
	}

	os.Setenv("PATH", os.Getenv("PATH")+":"+extrapath)
}

func init_suicide() {
	if Config.Suicide.Enabled {
		if Config.Suicide.Command == "" {
			Config.Suicide.Command = "sudo shutdown -h now"
		}

		if Config.Suicide.Time == 0 {
			Config.Suicide.Time = 600
		}
	}
}
