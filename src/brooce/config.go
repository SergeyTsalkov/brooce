package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"brooce/myip"

	redis "gopkg.in/redis.v3"
)

var redisClient *redis.Client
var config configType
var myProcName string
var logger = log.New(os.Stdout, "", log.LstdFlags)

type configType struct {
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

func init_config() {
	homedir := os.Getenv("HOME")
	bytes, err := ioutil.ReadFile(filepath.Join(homedir, ".brooce"))
	if err != nil {
		logger.Fatalln(err)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		logger.Fatalln(err)
	}

	init_redis()
	init_procname()
	init_syslog()

	init_path()
	init_suicide()
}

func init_redis() {
	if !strings.Contains(config.Redis.Host, ":") {
		config.Redis.Host = config.Redis.Host + ":6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         config.Redis.Host,
		Password:     config.Redis.Password,
		MaxRetries:   10,
		PoolSize:     10,
		DialTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolTimeout:  30 * time.Second,
	})
}

func init_procname() {
	ip := myip.PublicIPv4()
	if ip == "" {
		log.Fatalln("Unable to determine our IPv4 address!")
	}

	myProcName = fmt.Sprintf("%v-%v", ip, os.Getpid())
}

func init_path() {
	if config.Path == "" {
		return
	}

	extrapath := config.Path
	if !strings.HasPrefix(extrapath, "/") {
		extrapath = filepath.Join(os.Getenv("HOME"), extrapath)
	}

	os.Setenv("PATH", os.Getenv("PATH")+":"+extrapath)
}

func init_suicide() {
	if config.Suicide.Enabled {
		if config.Suicide.Command == "" {
			config.Suicide.Command = "sudo shutdown -h now"
		}

		if config.Suicide.Time == 0 {
			config.Suicide.Time = 600
		}

		logger.Println("After", config.Suicide.Time, "seconds of inactivity, we will run:", config.Suicide.Command)
	}
}
