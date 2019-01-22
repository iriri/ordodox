package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/iriri/minimal/flag"
)

type Board struct{ Name, Title string }

type Opt struct {
	Db     string
	ErrLog string
	Log    string
	Port   string
	Cache  string
	Domain string
	Boards []Board
}

func parseFlags() (*Opt, string) {
	var opt Opt
	var path string
	flag.String(&path, 'c', "", "ordodox.toml", "config file")
	flag.String(&opt.Db, 'd', "", "ordodox.db", "db file")
	flag.String(&opt.ErrLog, 'e', "", "", "error log file")
	flag.String(&opt.Log, 'f', "", "ordodox.log", "server log file")
	flag.String(&opt.Cache, 'g', "", "ordodox-autocert", "autocert cache dir")
	flag.Parse(1)
	return &opt, path
}

func parseConf(opt *Opt, path string) (*Opt, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	md, err := toml.DecodeFile(path, opt)
	if err != nil {
		panic(err)
	}
	if ud := md.Undecoded(); len(ud) > 0 {
		return nil, fmt.Errorf("unexpected key(s) in config file: %v", ud)
	}
	return opt, nil
}

func Parse() (*Opt, error) {
	opt, path := parseFlags()
	return parseConf(opt, path)
}
