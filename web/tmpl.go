package web

import (
	"brooce/config"
	"brooce/util"
	"embed"
	"html/template"
	"log"
	"strings"
	"time"
)

//go:embed tmpl
var templateContent embed.FS

func makeTemplate() *template.Template {
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
		"BasePath": func() string {
			return config.Config.Web.BasePath
		},
		"TimeSince": func(timestamp int64) string {
			if timestamp == 0 {
				return ""
			}
			return util.HumanDuration(time.Since(time.Unix(timestamp, 0)), 1) + " ago"
		},
		"TimeBetween": func(start, end int64) string {
			if start == 0 || end == 0 {
				return ""
			}
			if start > end {
				start, end = end, start
			}

			return util.HumanDuration(time.Unix(end, 0).Sub(time.Unix(start, 0)), 1)
		},
		"TimeDuration": func(seconds int) string {
			if seconds == 0 {
				return ""
			}
			return util.HumanDuration(time.Duration(seconds)*time.Second, 1)
		},
		"FormatTime": func(timestamp int64) string {
			if timestamp == 0 {
				return ""
			}
			return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		},
		"CurrentTime": func() string {
			return time.Now().UTC().Format("2006-01-02 15:04:05")
		},
		"Join": func(slice []string, connector string) string {
			return strings.Join(slice, connector)
		},
	})

	_, err := tpl.ParseFS(templateContent, "tmpl/*.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return tpl
}
