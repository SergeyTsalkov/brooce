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

var BrooceDir = filepath.Join(os.Getenv("HOME"), ".brooce")

type ConfigType struct {
	ClusterName string `json:"cluster_name"`
	ProcName    string `json:"-"`

	Timeout int `json:"timeout"`

	Web struct {
		Addr     string `json:"addr"`
		CertFile string `json:"certfile"`
		KeyFile  string `json:"keyfile"`
		Username string `json:"username"`
		Password string `json:"password"`
		NoAuth   bool   `json:"no_auth"`
		Disable  bool   `json:"disable"`
	} `json:"web"`

	FileOutputLog struct {
		Enable bool `json:"enable"`
	} `json:"file_output_log"`

	RedisOutputLog struct {
		DropDone    bool  `json:"drop_done"`
		DropFailed  bool  `json:"drop_failed"`
		ExpireAfter int64 `json:"expire_after"`
	} `json:"redis_output_log"`

	JobResults struct {
		DropDone   bool `json:"drop_done"`
		DropFailed bool `json:"drop_failed"`
	} `json:"job_results"`

	Redis struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`

	Suicide struct {
		Enable  bool   `json:"enable"`
		Command string `json:"command"`
		Time    int    `json:"time"`
	} `json:"suicide"`

	Queues map[string]int `json:"queues"`
	Path   string         `json:"path"`
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
	if !util.IsDir(BrooceDir) {
		err := os.Mkdir(BrooceDir, 0755)
		if err != nil {
			log.Fatalln("Unable to create directory", BrooceDir, ":", err)
		}
	}

	configFile := filepath.Join(BrooceDir, "brooce.conf")

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read config file", configFile, "so using defaults!")
	} else {
		err = json.Unmarshal(bytes, &Config)
		if err != nil {
			log.Fatalln("Your config file", configFile, "seem to have invalid json! Please fix it or delete the file!")
		}
	}

	init_defaults()

	if !util.FileExists(configFile) {
		if bytes, err := json.MarshalIndent(&Config, "", "  "); err == nil {
			err = ioutil.WriteFile(configFile, bytes, 0744)
			if err != nil {
				log.Println("Warning: Unable to write clean config file to", configFile, ", error was:", err)
			} else {
				log.Println("We wrote a default config file to", configFile)
			}
		}
	}
}

func init_defaults() {
	if Config.ClusterName == "" {
		Config.ClusterName = "brooce"
	}

	if Config.ProcName == "" {
		Config.ProcName = fmt.Sprintf("%v-%v", myip.PublicIPv4(), os.Getpid())
	}

	if !Config.Web.Disable {
		if Config.Web.Addr == "" {
			Config.Web.Addr = ":8080"
		}

		if !Config.Web.NoAuth && (Config.Web.Username == "" || Config.Web.Password == "") {
			Config.Web.Username = "admin"
			Config.Web.Password = util.RandomString(8)
			log.Printf("You didn't specify a web username/password, so we generated these: %s/%s", Config.Web.Username, Config.Web.Password)
		}

		Config.Web.CertFile = cleanpath(Config.Web.CertFile)
		Config.Web.KeyFile = cleanpath(Config.Web.KeyFile)
	}

	if Config.RedisOutputLog.ExpireAfter == 0 {
		Config.RedisOutputLog.ExpireAfter = 604800 // 7 days
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

	if Config.Suicide.Enable {
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

func cleanpath(path string) string {
	if path == "" || strings.HasPrefix(path, "/") {
		return path
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(path, "~/"))
	}

	return filepath.Join(BrooceDir, path)
}
