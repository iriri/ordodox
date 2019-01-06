package main

import (
	"net/http"

	"ordodox/config"
	"ordodox/database"
	"ordodox/server"
)

func main() {
	opt, err := config.Parse()
	if err != nil {
		panic(err)
	}
	if err = database.Init(opt.Db, opt.Boards); err != nil {
		panic(err)
	}
	if err = http.ListenAndServe(opt.Port, server.New(opt.Boards, opt.Log)); err != nil {
		panic(err)
	}
}
