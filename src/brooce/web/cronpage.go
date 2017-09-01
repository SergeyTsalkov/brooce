package web

import (
	"fmt"
	"log"
	"net/http"

	"brooce/cron"
	"brooce/listing"
)

type cronpageOutputType struct {
	Crons         map[string]*cron.CronType
	DisabledCrons map[string]*cron.CronType
	Edit          string
	New           bool
}

func cronpageHandler(req *http.Request, rep *httpReply) (err error) {
	output := &cronpageOutputType{}

	output.Crons, err = listing.Crons()
	if err != nil {
		return
	}
	output.DisabledCrons, err = listing.DisabledCrons()
	if err != nil {
		return
	}

	//output.Edit = req.FormValue("edit")
	//output.New = (req.FormValue("new") == "1")

	err = templates.ExecuteTemplate(rep, "cronpage", output)
	return
}

func deleteCronHandler(req *http.Request, rep *httpReply) (err error) {
	if item := req.FormValue("item"); item != "" {
		enabledKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		disabledKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		err = redisClient.Del(enabledKey, disabledKey).Err()
	}

	return
}

func disableCronHandler(req *http.Request, rep *httpReply) (err error) {
	if item := req.FormValue("item"); item != "" {
		srcKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		dstKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		err = redisClient.Rename(srcKey, dstKey).Err()
	}

	return
}

func enableCronHandler(req *http.Request, rep *httpReply) (err error) {
	if item := req.FormValue("item"); item != "" {
		srcKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		dstKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		err = redisClient.Rename(srcKey, dstKey).Err()
	}

	return
}

func scheduleCronHandler(req *http.Request, rep *httpReply) (err error) {
	if name := req.FormValue("item"); name != "" {
		// log.Printf("GOT ITEM NAMED %+v", name)
		keyPrefix := fmt.Sprintf("%s:cron:jobs", redisHeader)
		val := redisClient.Get(fmt.Sprintf("%s:%s", keyPrefix, name)).Val()

		job, err := cron.ParseCronLine(name, val)
		if err != nil {
			log.Printf("Error loading cron named %s with val: %s", name, val)
			return err
		}

		// log.Printf("Can schedule cron %s on queue %s with command: %s", name, job.Queue, job.Task().Json())
		// pendingList := strings.Join([]string{redisHeader, "queue", job.Queue, "pending"}, ":")
		pendingList := fmt.Sprintf("%s:queue:%s:pending", redisHeader, job.Queue)
		redisClient.LPush(pendingList, job.Task().Json())
	}
	return
}

/*
func saveCronHandler(req *http.Request, rep *httpReply) (err error) {
	name := req.FormValue("name")
	item := req.FormValue("item")

	_, err = cron.ParseCronLine(name, item)
	if err != nil {
		return
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		keyPrefix := fmt.Sprintf("%s:cron:jobs", redisHeader)
		if req.FormValue("disabled") == "true" {
			keyPrefix = fmt.Sprintf("%s:cron:disabledjobs", redisHeader)
		}

		oldname := req.FormValue("oldname")
		oldKey := fmt.Sprintf("%s:%s", keyPrefix, oldname)
		newKey := fmt.Sprintf("%s:%s", keyPrefix, name)

		pipe.Del(oldKey)
		pipe.Set(newKey, item, 0)
		return nil
	})

	return
}
*/
