package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func (srv *server) appAdded(wr http.ResponseWriter, req *http.Request) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))
}
