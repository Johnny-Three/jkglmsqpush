package httpserver

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{

	Route{
		"TodoShow",
		"GET",
		"/v1/jkglmsgpush_api/{userid}",
		TodoShowTask,
	},
}
