package tpl

import (
	"html/template"
	"log"

	"brooce/config"
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
	})

	for _, tplString := range tplList {
		_, err := tpl.Parse(tplString)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return tpl
}
