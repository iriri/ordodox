package main

import (
	"log"
	"net/http"

	"gopkg.in/natefinch/lumberjack.v2"

	"ordodox/config"
	"ordodox/database"
	"ordodox/server"
)

func main() {
	opt, err := config.Parse()
	if err != nil {
		panic(err)
	}
	if opt.ErrLog != "" {
		log.SetFlags(log.LstdFlags | log.LUTC)
		log.SetOutput(&lumberjack.Logger{
			Filename:   opt.ErrLog,
			MaxSize:    128,
			MaxBackups: 10,
			Compress:   true,
		})
	}
	if err = database.Init(opt); err != nil {
		panic(err)
	}

	serve, done := server.Init(opt)
	log.Println("serving")
	if err = serve(); err != http.ErrServerClosed {
		panic(err)
	}
	<-done
	log.Println("done serving")
}
