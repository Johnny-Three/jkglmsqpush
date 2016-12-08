package httpserver

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"wbproject/jkglmsgpush/src/user"

	"github.com/gorilla/mux"
)

var tmp *user.Users

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func StoreUsers(u *user.Users) {
	tmp = u
}

func TodoShowTask(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userid, err := strconv.Atoi(vars["userid"])
	if err != nil {
		panic(err)
	}

	//back info ..
	var result Result
	w.Header().Set("Content-Type", "text/html")
	result.Data = strings.Replace(tmp.ToString(userid), "\n", "<br>", -1)
	io.WriteString(w, result.Data)
}
