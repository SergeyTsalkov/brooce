package web

import (
	"log"
	"net/http"
	"text/template"
	"time"
)

var reqHandler = http.NewServeMux()

var serv = &http.Server{
	Addr:           ":8080",
	Handler:        reqHandler,
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20,
}

func Start() {
	tpl := template.New("")

	for _, tplString := range tplList {
		_, err := tpl.Parse(tplString)
		if err != nil {
			log.Fatalln(err)
		}
	}

	reqHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := tpl.Execute(w, map[string]string{})
		if err != nil {
			log.Println(err)
		}
	})

	err := serv.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
