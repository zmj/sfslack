package main

import "net/http"

func (srv *server) newCommand(wr http.ResponseWriter, req *http.Request) {
	// parse cmd
	// if not parse
	//   print help
	// parse user
	// check auth cache
	// if cache hit
	//   init workflow with first msg callback
	// else
	//   auth prompt
	//   init workflow with no callback
}
