package database

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"unicode"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

func lookupRef(id int64) string {
	return ""
}

func flushGreen(res, sub *strings.Builder) {
	res.WriteString("<span class=\"gt\">")
	res.WriteString(html.EscapeString(sub.String()))
	res.WriteString("</span>")
	sub.Reset()
}

func flushRef(conn *sqlite3.Conn, board string, res, sub *strings.Builder, op int64) error {
	id, err := strconv.ParseInt(sub.String()[2:], 10, 64)
	if err != nil {
		res.WriteString(html.EscapeString(sub.String()))
		sub.Reset()
		return nil
	}

	stmt, err := conn.Prepare(fmt.Sprintf("SELECT op FROM %s_posts WHERE id = %d", board, id))
	if err != nil {
		return err
	}
	defer stmt.Close()
	if ok, err := stmt.Step(); err != nil {
		return err
	} else if !ok {
		res.WriteString(html.EscapeString(sub.String()))
		sub.Reset()
		return nil
	}
	var op_ int64
	err = stmt.Scan(&op_)
	if err != nil {
		return err
	}

	if op_ == op {
		res.WriteString(fmt.Sprintf("<a href=\"#%d\">", id))
	} else {
		res.WriteString(fmt.Sprintf("<a href=\"/%s/%d#%d\">", board, op_, id))
	}
	res.WriteString(html.EscapeString(sub.String()))
	res.WriteString("</a>")
	sub.Reset()
	return nil
}

func parse(conn *sqlite3.Conn, board string, comm interface{}, op int64) (interface{}, error) {
	comm_, ok := comm.(string)
	if !ok {
		// assert(comm == nil) (^:
		return nil, nil
	}

	type token int
	const (
		plain token = iota
		gt
		green
		ref
	)
	state := plain
	res, sub := new(strings.Builder), new(strings.Builder)
	for _, c := range comm_ {
		switch c {
		case '\r':
			switch state {
			case gt:
				state = green
			case ref:
				if err := flushRef(conn, board, res, sub, op); err != nil {
					return nil, err
				}
				state = plain
			case plain, green:
			}
		case '\n':
			switch state {
			case plain:
				res.WriteString(html.EscapeString(sub.String()))
				sub.Reset()
			case gt, green:
				flushGreen(res, sub)
			case ref:
				if err := flushRef(conn, board, res, sub, op); err != nil {
					return nil, err
				}
			}
			state = plain
			res.WriteString("<br>")
		case '>':
			switch state {
			case plain:
				res.WriteString(html.EscapeString(sub.String()))
				sub.Reset()
				state = gt
			case gt:
				state = ref
			case ref:
				if err := flushRef(conn, board, res, sub, op); err != nil {
					return nil, err
				}
				state = plain
			case green:
			}
			sub.WriteRune(c)
		default:
			switch state {
			case gt:
				state = green
			case ref:
				if !unicode.IsDigit(c) {
					if err := flushRef(conn, board, res, sub, op); err != nil {
						return nil, err
					}
					state = plain
				}
			case plain, green:
			}
			sub.WriteRune(c)
		}
	}
	switch state {
	case plain:
		res.WriteString(html.EscapeString(sub.String()))
	case gt, green:
		flushGreen(res, sub)
	case ref:
		if err := flushRef(conn, board, res, sub, op); err != nil {
			return nil, err
		}
	}
	return res.String(), nil
}
