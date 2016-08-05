package web

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"brooce/config"
	myredis "brooce/redis"
	"brooce/web/tpl"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

var reqHandler = http.NewServeMux()
var templates = tpl.Get()

var serv = &http.Server{
	Addr:         config.Config.Web.Addr,
	Handler:      reqHandler,
	ReadTimeout:  10 * time.Second,
	WriteTimeout: 10 * time.Second,
}

func Start() {
	if config.Config.Web.Username == "" || config.Config.Web.Password == "" {
		log.Println("Web interface disabled -- you didn't configure username/password!")
		return
	}

	reqHandler.HandleFunc("/", makeHandler(mainpageHandler, "GET"))
	reqHandler.HandleFunc("/failed/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/done/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/delayed/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/pending/", makeHandler(joblistHandler, "GET"))

	reqHandler.HandleFunc("/retry/failed/", makeHandler(retryHandler, "POST"))
	reqHandler.HandleFunc("/retry/done/", makeHandler(retryHandler, "POST"))
	reqHandler.HandleFunc("/retry/delayed/", makeHandler(retryHandler, "POST"))

	reqHandler.HandleFunc("/retryall/failed/", makeHandler(retryAllHandler, "POST"))
	reqHandler.HandleFunc("/retryall/delayed/", makeHandler(retryAllHandler, "POST"))

	reqHandler.HandleFunc("/delete/failed/", makeHandler(deleteHandler, "POST"))
	reqHandler.HandleFunc("/delete/done/", makeHandler(deleteHandler, "POST"))
	reqHandler.HandleFunc("/delete/delayed/", makeHandler(deleteHandler, "POST"))
	reqHandler.HandleFunc("/delete/pending/", makeHandler(deleteHandler, "POST"))

	reqHandler.HandleFunc("/deleteall/failed/", makeHandler(deleteAllHandler, "POST"))
	reqHandler.HandleFunc("/deleteall/done/", makeHandler(deleteAllHandler, "POST"))
	reqHandler.HandleFunc("/deleteall/delayed/", makeHandler(deleteAllHandler, "POST"))
	reqHandler.HandleFunc("/deleteall/pending/", makeHandler(deleteAllHandler, "POST"))

	reqHandler.HandleFunc("/showlog/", makeHandler(showlogHandler, "GET"))

	go func() {
		log.Println("Web server listening on", config.Config.Web.Addr)
		err := serv.ListenAndServe()
		if err != nil {
			log.Fatalln(err)
		}
	}()
}

func makeHandler(fn func(*http.Request) (*bytes.Buffer, error), method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if r := recover(); r != nil {
				str := fmt.Sprintf("%v", r)
				log.Println("Recovered from panic:", str)
				http.Error(w, str, http.StatusInternalServerError)
				return
			}
		}()

		username, password, authOk := r.BasicAuth()
		if !authOk || username != config.Config.Web.Username || password != config.Config.Web.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="brooce"`)
			http.Error(w, "401 Unauthorized\n", http.StatusUnauthorized)
			return
		}

		if method != r.Method || r.URL.Path == "/favicon.ico" {
			http.Error(w, "404 File Not Found\n", http.StatusNotFound)
			return
		}

		if method != "GET" && r.FormValue("csrf") != config.Config.CSRF() {
			http.Error(w, "Invalid csrf value!\n", http.StatusUnauthorized)
			return
		}

		buf, err := fn(r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if buf == nil {
			buf = &bytes.Buffer{}
		}

		if method != "GET" && buf.Len() == 0 {
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		io.Copy(w, buf)
	}
}
