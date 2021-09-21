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

var BrooceConfigDir = os.Getenv("BROOCE_CONFIGDIR")
var BrooceLogDir = os.Getenv("BROOCE_LOGDIR")

type ConfigType struct {
	ClusterName    string `json:"cluster_name"`
	ClusterLogName string `json:"-"`
	ProcName       string `json:"-"`

	GlobalJobOptions JobOptions `json:"global_job_options"`

	Web struct {
		Addr     string `json:"addr"`
		BasePath string `json:"basepath"`
		CertFile string `json:"certfile"`
		KeyFile  string `json:"keyfile"`
		Username string `json:"username"`
		Password string `json:"password"`
		NoAuth   bool   `json:"no_auth"`
		NoLog    bool   `json:"no_log"`
		Disable  bool   `json:"disable"`
	} `json:"web"`

	FileOutputLog struct {
		Enable bool `json:"enable"`
	} `json:"file_output_log"`

	Redis struct {
		Host     string `json:"host"`
		Socket   string `json:"socket"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`

	Suicide struct {
		Enable  bool   `json:"enable"`
		Command string `json:"command"`
		Time    int    `json:"time"`
	} `json:"suicide"`

	Queues []Queue `json:"queues"`

	Path string `json:"path"`
}

type Queue struct {
	Name       string `json:"name"`
	Workers    int    `json:"workers"`
	JobOptions `json:"job_options"`
}

var Config = ConfigType{}

func (c *ConfigType) CSRF() string {
	return util.Md5sum(c.Web.Username + ":" + c.Web.Password)
}

func (c *ConfigType) JobOptionsForQueue(queue string) (opts JobOptions) {
	for _, q := range c.Queues {
		if q.Name == queue {
			opts = q.JobOptions
			return
		}
	}

	return
}

func (q *Queue) DeepJobOptions() (j JobOptions) {
	j.Merge(q.JobOptions)
	j.Merge(Config.GlobalJobOptions)
	j.Merge(DefaultJobOptions)
	return
}

func (q *Queue) PendingList() string {
	return fmt.Sprintf("%s:queue:%s:pending", Config.ClusterName, q.Name)
}

func (q *Queue) DoneList() string {
	return fmt.Sprintf("%s:queue:%s:done", Config.ClusterName, q.Name)
}

func (q *Queue) FailedList() string {
	return fmt.Sprintf("%s:queue:%s:failed", Config.ClusterName, q.Name)
}

func (q *Queue) DelayedList() string {
	return fmt.Sprintf("%s:queue:%s:delayed", Config.ClusterName, q.Name)
}

// this use of init sucks, but we'll have to fix every "var redisClient = myredis.Get()"
// that is in a header to avoid it -- let's do this later!
func init() {

	initDefaultDirs()

	configFile := filepath.Join(BrooceConfigDir, "brooce.conf")
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read config file", configFile, "so using defaults!")
	} else {
		err = json.Unmarshal(bytes, &Config)
		if err != nil {
			log.Fatalln("Your config file", configFile, "seem to have invalid json! Please fix it or delete the file!")
		}
	}

	if BrooceConfigDir != "" {
		log.Println("ConfigDir:", BrooceConfigDir)
	}
	if BrooceLogDir != "" {
		log.Println("LogDir:", BrooceLogDir)
	}

	initDefaultJobOptions()
	initDefaultConfig()

	if !util.FileExists(configFile) {
		if bytes, err := json.MarshalIndent(&Config, "", "  "); err == nil {
			err = ioutil.WriteFile(configFile, bytes, 0644)
			if err != nil {
				log.Println("Warning: Unable to write default config file to", configFile, ", error was:", err)
			} else {
				log.Println("We wrote a default config file to", configFile)
			}
		}
	}

	initThreads()
}

func initDefaultDirs() {

	if BrooceConfigDir == "" {
		BrooceConfigDir = filepath.Join(os.Getenv("HOME"), ".brooce")
	}

	if !util.IsDir(BrooceConfigDir) {
		err := os.Mkdir(BrooceConfigDir, 0755)
		if err != nil {
			log.Fatalln("Unable to create directory", BrooceConfigDir, ":", err)
		}
	}

	if BrooceLogDir == "" {
		BrooceLogDir = filepath.Join(BrooceConfigDir, "logs")
	}

	// for backward compatibility, but you can create it instead, or remove this, if you wish
	if !util.IsDir(BrooceLogDir) {
		BrooceLogDir = BrooceConfigDir
	}

}

func initDefaultConfig() {
	if Config.ClusterName == "" {
		Config.ClusterName = "brooce"
	}

	if Config.ClusterLogName == "" {
		Config.ClusterLogName = strings.Replace(Config.ClusterName, "{", "", 1)
		Config.ClusterLogName = strings.Replace(Config.ClusterLogName, "}", "", 1)
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

		if Config.Web.BasePath != "" {
			Config.Web.BasePath = strings.Trim(Config.Web.BasePath, "/")
			Config.Web.BasePath = "/" + Config.Web.BasePath
		}

		Config.Web.CertFile = cleanpath(Config.Web.CertFile)
		Config.Web.KeyFile = cleanpath(Config.Web.KeyFile)
	}

	if Config.Queues == nil {
		Config.Queues = []Queue{
			{
				Name:    "common",
				Workers: 1,
			},
		}
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

		if Config.Suicide.Time <= 0 {
			Config.Suicide.Time = 600
		}
	}

	if Config.Path != "" {
		os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+Config.Path)
	}

	if Config.GlobalJobOptions == (JobOptions{}) {
		Config.GlobalJobOptions = DefaultJobOptions
	}
}

func cleanpath(path string) string {
	if path == "" || strings.HasPrefix(path, string(os.PathSeparator)) {
		return path
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(path, "~/"))
	}

	return filepath.Join(BrooceConfigDir, path)
}
