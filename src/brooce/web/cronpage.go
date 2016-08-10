package web

import (
	"bytes"
	"fmt"
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

func cronpageHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
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

	err = templates.ExecuteTemplate(buf, "cronpage", output)
	return
}

func deleteCronHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	if item := req.FormValue("item"); item != "" {
		enabledKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		disabledKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		err = redisClient.Del(enabledKey, disabledKey).Err()
	}

	return
}

func disableCronHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	if item := req.FormValue("item"); item != "" {
		srcKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		dstKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		err = redisClient.Rename(srcKey, dstKey).Err()
	}

	return
}

func enableCronHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	if item := req.FormValue("item"); item != "" {
		srcKey := fmt.Sprintf("%s:cron:disabledjobs:%s", redisHeader, item)
		dstKey := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		err = redisClient.Rename(srcKey, dstKey).Err()
	}

	return
}

/*
func saveCronHandler(req *http.Request) (buf *bytes.Buffer, err error) {
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
