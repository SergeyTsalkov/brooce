package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"brooce/config"
	myredis "brooce/redis"
	"brooce/web/tpl"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName
var redisLogHeader = config.Config.ClusterLogName

var reqHandler = http.NewServeMux()
var templates = tpl.Get()

var serv = &http.Server{
	Addr:         config.Config.Web.Addr,
	Handler:      reqHandler,
	ReadTimeout:  10 * time.Second,
	WriteTimeout: 10 * time.Second,
}

var webLogWriter *os.File

type httpReply struct {
	*bytes.Buffer
	contentType string
	statusCode int
	redirect   string
}

type httpHandler func(*http.Request, *httpReply) error
type middleware func(httpHandler) httpHandler

func Start() {
	if config.Config.Web.Disable {
		log.Println("Web interface disabled!")
		return
	}

	reqHandler.HandleFunc("/", makeHandler(mainpageHandler, "GET"))

	reqHandler.HandleFunc("/js/", makeHandler(resourceHandler, "GET"))
	reqHandler.HandleFunc("/css/", makeHandler(resourceHandler, "GET"))
	reqHandler.HandleFunc("/fonts/", makeHandler(resourceHandler, "GET"))

	reqHandler.HandleFunc("/failed/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/done/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/delayed/", makeHandler(joblistHandler, "GET"))
	reqHandler.HandleFunc("/pending/", makeHandler(joblistHandler, "GET"))

	reqHandler.HandleFunc("/search", makeHandler(searchHandler, "GET"))

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

	reqHandler.HandleFunc("/cron", makeHandler(cronpageHandler, "GET"))
	//reqHandler.HandleFunc("/savecron", makeHandler(saveCronHandler, "POST"))
	reqHandler.HandleFunc("/deletecron", makeHandler(deleteCronHandler, "POST"))
	reqHandler.HandleFunc("/disablecron", makeHandler(disableCronHandler, "POST"))
	reqHandler.HandleFunc("/enablecron", makeHandler(enableCronHandler, "POST"))
	reqHandler.HandleFunc("/schedulecron", makeHandler(scheduleCronHandler, "POST"))

	go func() {
		var err error

		if !config.Config.Web.NoLog {
			filename := filepath.Join(config.BrooceLogDir, "web.log")
			webLogWriter, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalln("Unable to open logfile", filename, "for writing! Error was", err)
			}
			defer webLogWriter.Close()
		}

		if config.Config.Web.CertFile == "" && config.Config.Web.KeyFile == "" {
			log.Println("Starting HTTP server on", config.Config.Web.Addr)
			err = serv.ListenAndServe()
		} else {
			log.Println("Starting HTTPS server on", config.Config.Web.Addr)
			err = serv.ListenAndServeTLS(config.Config.Web.CertFile, config.Config.Web.KeyFile)
		}

		if err != nil {
			log.Println("Warning: Unable to start web server, error was:", err)
		}
	}()
}

func makeHandler(fn httpHandler, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		rep := &httpReply{
			Buffer:     &bytes.Buffer{},
			statusCode: http.StatusOK,
		}

		var err error
		if method == req.Method {
			err = allMiddleware(fn)(req, rep)
		} else {
			rep.statusCode = http.StatusNotFound
		}

		if req.Method != "GET" && rep.statusCode == http.StatusOK && rep.Len() == 0 {
			rep.statusCode = http.StatusSeeOther
		}

		repLength := rep.Len()

		if (rep.contentType != "") {
			w.Header().Set("Content-Type", rep.contentType)
		}

		switch rep.statusCode {
		case http.StatusInternalServerError:
			http.Error(w, "500 Internal Server Error\n", rep.statusCode)
		case http.StatusNotFound:
			http.Error(w, "404 File Not Found\n", rep.statusCode)
		case http.StatusForbidden:
			http.Error(w, "403 Forbidden\n", rep.statusCode)
		case http.StatusUnauthorized:
			w.Header().Set("WWW-Authenticate", `Basic realm="brooce"`)
			http.Error(w, "401 Password Required\n", rep.statusCode)
		case http.StatusSeeOther:
			if rep.redirect == "" {
				rep.redirect = req.Referer()
			}
			http.Redirect(w, req, rep.redirect, http.StatusSeeOther)
		case http.StatusNotModified:
			w.Header().Set("Expires", time.Now().AddDate(0, 0, 1).Format(http.TimeFormat))
			rep.statusCode = http.StatusOK
			io.Copy(w, rep)
		default:
			io.Copy(w, rep)
		}

		if webLogWriter != nil {
			logLine := fmt.Sprintf(`%s - [%s] "%s %s" %d %d "%s" "%s"`,
				req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")],
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				req.Method,
				req.RequestURI,
				rep.statusCode,
				repLength,
				req.Referer(),
				req.UserAgent(),
			)

			webLogWriter.Write([]byte(logLine + "\n"))
			if err != nil {
				webLogWriter.Write([]byte(fmt.Sprintf("ERROR IN LAST REQUEST: %v\n", err)))
			}
		}

		if err != nil {
			log.Println("[web] Error:", err)
		}

	}
}

func allMiddleware(fn httpHandler) httpHandler {
	// these middlewares will be run in reverse!
	middlewares := []middleware{
		csrfMiddleware,
		authMiddleware,
		errorMiddleware,
	}

	allMiddlewareFn := fn

	for _, middlewareFunc := range middlewares {
		allMiddlewareFn = middlewareFunc(allMiddlewareFn)
	}
	return allMiddlewareFn
}

func authMiddleware(next httpHandler) httpHandler {
	return func(req *http.Request, rep *httpReply) (err error) {

		if config.Config.Web.Username != "" && config.Config.Web.Password != "" {
			username, password, authOk := req.BasicAuth()
			if !authOk || username != config.Config.Web.Username || password != config.Config.Web.Password {
				rep.statusCode = http.StatusUnauthorized
				return
			}
		}

		err = next(req, rep)
		return
	}
}

func errorMiddleware(next httpHandler) httpHandler {
	return func(req *http.Request, rep *httpReply) (err error) {

		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("Panic: %v", r)
				rep.statusCode = http.StatusInternalServerError
				return
			}
		}()

		err = next(req, rep)

		if err != nil {
			rep.statusCode = http.StatusInternalServerError
		}

		return
	}
}

func csrfMiddleware(next httpHandler) httpHandler {
	return func(req *http.Request, rep *httpReply) (err error) {

		if req.Method != "GET" && req.FormValue("csrf") != config.Config.CSRF() {
			rep.statusCode = http.StatusForbidden
			return
		}

		err = next(req, rep)
		return
	}
}

func resourceHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 2 {
		err = fmt.Errorf("Invalid path")
		return
	}

	resFileType := path[0]
	resFileName := path[1]
	if config.BrooceResourceDir != "" {
		resFilePath := filepath.Join(config.BrooceResourceDir, filepath.Base(resFileName))
		if _, err = os.Stat(resFilePath); err == nil {
			var content []byte
			content, err = ioutil.ReadFile(resFilePath)
			if err != nil {
				rep.statusCode = http.StatusInternalServerError
				return
			}
			rep.Buffer = bytes.NewBuffer(content)

			switch resFileType {
			case "js":
				rep.contentType = "application/javascript"
			case "css":
				rep.contentType = "text/css"
			case "fonts":
				rep.contentType = "font/woff2"
			}

			rep.statusCode = http.StatusNotModified
			return
		} else {
			err = nil
		}
	}

	rep.statusCode = http.StatusSeeOther
	switch resFileName {
	case "glyphicons-halflings-regular.woff2":
		rep.redirect = "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/fonts/glyphicons-halflings-regular.woff2"
	case "bootstrap.min.css":
		rep.redirect = "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css"
	case "bootstrap.min.js":
		rep.redirect = "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js"
	case "html5shiv.min.js":
		rep.redirect = "https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"
	case "respond.min.js":
		rep.redirect = "https://oss.maxcdn.com/respond/1.4.2/respond.min.js"
	case "jquery-2.2.4.min.js":
		rep.redirect = "https://code.jquery.com/jquery-2.2.4.min.js"
	default:
		rep.statusCode = http.StatusNotFound
	}

	return
}
