package main

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"time"
)

type tsWriter struct {
	w io.Writer
}

func (writer *tsWriter) Write(p []byte) (int, error) {
	ts := time.Now().Format("2006/01/02 15:04:05")
	return writer.w.Write([]byte(ts + " " + string(p)))
}

func init_syslog() {
	if config.Syslog.Host != "" {
		syslogWriter, err := syslog.Dial("udp", config.Syslog.Host, syslog.LOG_ALERT|syslog.LOG_USER, "brooce")
		if err != nil {
			log.Fatalln(err)
		}

		stdoutWriter := &tsWriter{os.Stdout}

		prefix := fmt.Sprintf("[%v] ", myProcName)
		logger = log.New(io.MultiWriter(syslogWriter, stdoutWriter), prefix, 0)
	}

}
