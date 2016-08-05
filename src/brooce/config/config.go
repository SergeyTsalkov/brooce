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
	"brooce/util"
)

type ConfigType struct {
	ClusterName string `json:"cluster_name"`
	ProcName    string `json:"process_name"`

	Timeout int

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

var Config = ConfigType{}

func (c *ConfigType) TotalThreads() (threads int) {
	for _, ct := range Config.Queues {
		threads += ct
	}
	return
}

func (c *ConfigType) CSRF() string {
	return util.Md5sum(c.Web.Username + ":" + c.Web.Password)
}

func init() {
	configFile := filepath.Join(os.Getenv("HOME"), ".brooce")

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read config file", configFile, "so using defaults!")
	} else {
		err = json.Unmarshal(bytes, &Config)
		if err != nil {
			log.Println("Your config file", configFile, "seem to have invalid json! Using defaults instead!")
		}
	}

	init_defaults()
}

func init_defaults() {
	if Config.ClusterName == "" {
		Config.ClusterName = "brooce"
	}

	if Config.ProcName == "" {
		Config.ProcName = fmt.Sprintf("%v-%v", myip.PublicIPv4(), os.Getpid())
	}

	if Config.Web.Addr == "" {
		Config.Web.Addr = ":8080"
	}

	if Config.Web.Username == "" || Config.Web.Password == "" {
		Config.Web.Username = "admin"
		Config.Web.Password = util.RandomString(8)
		log.Printf("You didn't specify a web username/password, so we generated these: %s/%s", Config.Web.Username, Config.Web.Password)
	}

	if Config.Queues == nil {
		Config.Queues = map[string]int{"common": 1}
	}

	if Config.Timeout == 0 {
		Config.Timeout = 3600
	}

	if Config.Redis.Host == "" {
		Config.Redis.Host = "localhost"
	}

	if !strings.Contains(Config.Redis.Host, ":") {
		Config.Redis.Host = Config.Redis.Host + ":6379"
	}

	if Config.Suicide.Enabled {
		if Config.Suicide.Command == "" {
			Config.Suicide.Command = "sudo shutdown -h now"
		}

		if Config.Suicide.Time == 0 {
			Config.Suicide.Time = 600
		}
	}

	if Config.Path != "" {
		os.Setenv("PATH", os.Getenv("PATH")+":"+Config.Path)
	}
}
