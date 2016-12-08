package httpserver

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"
	. "wbproject/jkglmsgpush/src/enviroment"
)

func Loggeri(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		Logger.Info("request start:", r.Method, "\t", r.RequestURI, "\t", name, "\t",
			string(body))
		inner.ServeHTTP(w, r)
		Logger.Info("request end:", r.Method, "\t", r.RequestURI, "\t", name, "\t", time.Since(start), "\t",
			string(body))
	})
}
