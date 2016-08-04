package tpl

import (
	"html/template"
	"log"
	"time"

	"brooce/config"

	humanize "github.com/dustin/go-humanize"
)

var tplList = []string{
	headerTpl,
	footerTpl,
	mainPageTpl,
	jobListTpl,
}

func Get() *template.Template {
	tpl := template.New("")

	tpl.Funcs(template.FuncMap{
		"Iter": func(start, end int64) []int64 {
			result := make([]int64, end-start+1)
			for i := start; i <= end; i++ {
				result[i-start] = i
			}
			return result
		},
		"CSRF": func() string {
			return config.Config.CSRF()
		},
		"TimeSince": func(timestamp int64) string {
			return humanize.Time(time.Unix(timestamp, 0))
		},
	})

	for _, tplString := range tplList {
		_, err := tpl.Parse(tplString)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return tpl
}
