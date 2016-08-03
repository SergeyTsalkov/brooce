package tpl

import (
	"html/template"
	"log"
)

var tplList = []string{
	headerTpl,
	footerTpl,
	mainPageTpl,
	jobListTpl,
}

func Get() *template.Template {
	tpl := template.New("")

	for _, tplString := range tplList {
		_, err := tpl.Parse(tplString)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return tpl
}
