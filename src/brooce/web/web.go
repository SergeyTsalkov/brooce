package web

import (
	"bytes"
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

	reqHandler.HandleFunc("/", makeHandler(mainpageHandler))
	reqHandler.HandleFunc("/failed/", makeHandler(joblistHandler))
	reqHandler.HandleFunc("/done/", makeHandler(joblistHandler))
	reqHandler.HandleFunc("/delayed/", makeHandler(joblistHandler))
	reqHandler.HandleFunc("/pending/", makeHandler(joblistHandler))

	go func() {
		log.Println("Web server listening on", config.Config.Web.Addr)
		err := serv.ListenAndServe()
		if err != nil {
			log.Fatalln(err)
		}
	}()
}

func makeHandler(fn func(req *http.Request) (*bytes.Buffer, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, authOk := r.BasicAuth()
		if !authOk || username != config.Config.Web.Username || password != config.Config.Web.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="brooce"`)
			http.Error(w, "401 Unauthorized\n", http.StatusUnauthorized)
			return
		}

		buf, err := fn(r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(w, buf)
	}
}
