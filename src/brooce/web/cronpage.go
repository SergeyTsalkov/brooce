package web

import (
	"bytes"
	"net/http"

	"brooce/cron"
	"brooce/listing"
)

func cronpageHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}

	var crons map[string]*cron.CronType
	crons, err = listing.Crons()

	err = templates.ExecuteTemplate(buf, "cronpage", crons)
	return
}
