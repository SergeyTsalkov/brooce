package web

import (
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

	err = templates.ExecuteTemplate(rep, "cronpage", output)
	return
}

func deleteCronHandler(req *http.Request, rep *httpReply) error {
	job, err := cron.Get(req.FormValue("item"))
	if err != nil {
		return err
	}

	return job.Delete()
}

func disableCronHandler(req *http.Request, rep *httpReply) (err error) {
	job, err := cron.Get(req.FormValue("item"))
	if err != nil {
		return err
	}

	return job.Disable()
}

func enableCronHandler(req *http.Request, rep *httpReply) (err error) {
	job, err := cron.Get(req.FormValue("item"))
	if err != nil {
		return err
	}

	return job.Enable()
}

func scheduleCronHandler(req *http.Request, rep *httpReply) (err error) {
	job, err := cron.Get(req.FormValue("item"))
	if err != nil {
		return err
	}

	return job.Run()
}
