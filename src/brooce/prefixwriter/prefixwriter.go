package prefixwriter

import (
	"io"
	"strings"
	"time"
)

type PrefixWriter struct {
	io.Writer
	Prefix        string
	TsFormat      string
	firstLineDone bool
}

func (pw *PrefixWriter) Write(p []byte) (int, error) {
	prefix := pw.Prefix
	if strings.Contains(prefix, "--ts--") {
		prefix = strings.Replace(prefix, "--ts--", time.Now().Format(pw.TsFormat), -1)
	}

	str := strings.Replace(string(p), "\n", "\n"+prefix, -1)

	if !pw.firstLineDone {
		str = prefix + str
		pw.firstLineDone = true
	}

	_, err := pw.Writer.Write([]byte(str))

	// lie about how many bytes we wrote
	// if we don't, exec will barf
	return len(p), err
}
