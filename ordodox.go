package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/iriri/minimal/flag"
)

var opt struct {
	conf string
	db   string
	docs bool
	log  string
	port string
}

type board struct {
	Name  string
	Title string
}

func parseFlags() {
	flag.String(&opt.conf, 'c', "", "ordodox.toml", "config file")
	flag.String(&opt.db, 'd', "", "ordodox.db", "db file")
	flag.Bool(&opt.docs, 0, "docs", false, "generate docs")
	flag.String(&opt.log, 'l', "", "ordodox.log", "log file")
	flag.String(&opt.port, 'p', "", "8080", "listen port")
	flag.Parse(1)
}

func readConf() []board {
	if _, err := os.Stat(opt.conf); err != nil {
		panic(err)
	}
	
	var conf struct {
		Boards []board
	}
	if _, err := toml.DecodeFile(opt.conf, &conf); err != nil {
		panic(err)
	}
	return conf.Boards
}

func initDb(boards []board) {
	sq3, err := sqlite3.Open(opt.db)
	if err != nil {
		panic(err)
	}
	defer sq3.Close()

	err = sq3.Exec("CREATE TABLE IF NOT EXISTS boards(" +
		"name TEXT NOT NULL PRIMARY KEY," +
		"title TEXT NOT NULL) WITHOUT ROWID")
	if err != nil {
		panic(err)
	}

	insertStmt, err := sq3.Prepare("INSERT OR REPLACE INTO boards(name, title) VALUES(?, ?)")
	if err != nil {
		panic(err)
	}
	defer insertStmt.Close()

	for _, b := range boards {
		if err = insertStmt.Bind(b.Name, b.Title); err != nil {
			panic(err)
		}
		if _, err = insertStmt.Step(); err != nil {
			panic(err)
		}
		if err = insertStmt.Reset(); err != nil {
			panic(err)
		}
		err = sq3.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_posts("+
			"id INTEGER NOT NULL PRIMARY KEY,"+
			"op INTEGER NOT NULL,"+
			"ip INTEGER NOT NULL,"+
			"date DATETIME NOT NULL,"+
			"name TEXT,"+
			"email TEXT,"+
			"subject TEXT,"+
			"body TEXT,"+
			"file BLOB,"+
			"filename TEXT,"+
			"filesize INTEGER,"+
			"filewidth INTEGER,"+
			"fileheight INTEGER,"+
			"thumb BLOB,"+
			"thumbwidth INTEGER,"+
			"thumbheight INTEGER) WITHOUT ROWID", b.Name))
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	parseFlags()
	initDb(readConf())

	r := chi.NewRouter()
}
