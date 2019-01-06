package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/iriri/minimal/flag"
)

type Board struct{ Name, Title string }

type Opt struct {
	Db, Log, Port string
	Boards        []Board
}

func parseFlags() (*Opt, string) {
	var opt Opt
	var path string
	flag.String(&opt.Db, 'd', "", "ordodox.db", "db file")
	flag.String(&opt.Log, 'l', "", "ordodox.log", "log file")
	flag.String(&opt.Port, 'p', "", ":8080", "listen port")
	flag.String(&path, 'c', "", "ordodox.toml", "config file")
	flag.Parse(1)
	return &opt, path
}

func parseConf(path string) ([]Board, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	var conf struct{ Boards []Board }
	md, err := toml.DecodeFile(path, &conf)
	if err != nil {
		panic(err)
	}
	if ud := md.Undecoded(); len(ud) > 0 {
		return nil, fmt.Errorf("unexpected key(s) in config file: %v", ud)
	}
	return conf.Boards, nil
}

func Parse() (*Opt, error) {
	opt, path := parseFlags()
	var err error
	opt.Boards, err = parseConf(path)
	return opt, err
}
