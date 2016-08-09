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

	err = templates.ExecuteTemplate(buf, "cronpage", output)
	return
}

func deleteCronHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	if item := req.FormValue("item"); item != "" {
		key := fmt.Sprintf("%s:cron:jobs:%s", redisHeader, item)
		err = redisClient.Del(key).Err()
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
