package database

import (
	"fmt"

	"github.com/bvinc/go-sqlite-lite/sqlite3"

	"ordodox/config"
)

var path string
var boards = make(map[string]struct{})

func Init(path_ string, boards_ []config.Board) error {
	path = path_

	sq3, err := sqlite3.Open(path)
	if err != nil {
		return err
	}
	defer sq3.Close()
	err = sq3.Exec("CREATE TABLE IF NOT EXISTS boards(" +
		"name TEXT NOT NULL PRIMARY KEY," +
		"title TEXT NOT NULL) WITHOUT ROWID")
	if err != nil {
		return err
	}
	stmt, err := sq3.Prepare("INSERT OR REPLACE INTO boards VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range boards_ {
		boards[b.Name] = struct{}{}
		if err = stmt.Bind(b.Name, b.Title); err != nil {
			return err
		}
		if _, err = stmt.Step(); err != nil {
			return err
		}
		if err = stmt.Reset(); err != nil {
			return err
		}
		err = sq3.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_posts("+
			"id INTEGER PRIMARY KEY,"+
			"op INTEGER NOT NULL,"+
			"ip STRING NOT NULL,"+
			"date DATETIME NOT NULL,"+
			"name TEXT,"+
			"email TEXT,"+
			"subject TEXT,"+
			"body TEXT)", b.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

type BoardNotFound string

func (b BoardNotFound) Error() string {
	return fmt.Sprintf("Board not found: %s", b)
}

type OpNotFound int64

func (op OpNotFound) Error() string {
	return fmt.Sprintf("Op not found: %d", op)
}

type IdNotFound int64

func (id IdNotFound) Error() string {
	return fmt.Sprintf("Post not found: %d", id)
}

func validate(board string) (*sqlite3.Conn, error) {
	if _, ok := boards[board]; !ok {
		return nil, BoardNotFound(board)
	}
	return sqlite3.Open(path)
}

type Post struct {
	Id      int64
	Op      int64
	Ip      string
	Date    string
	Name    string
	Email   string
	Subject string
	Body    string
}

func Board(board string) ([][]Post, error) {
	sq3, err := validate(board)
	if err != nil {
		return nil, err
	}
	defer sq3.Close()

	stmt, err := sq3.Prepare(fmt.Sprintf("SELECT id FROM %s_posts "+
		"WHERE id = op "+
		"ORDER BY date DESC "+
		"LIMIT 10", board))
	if err != nil {
		return nil, err
	}
	threads := make([][]Post, 0, 10)
	for {
		if ok, err := stmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		var op int64
		err = stmt.Scan(&op)
		if err != nil {
			return nil, err
		}
		threads = append(threads, []Post{Post{Id: op}})
	}
	return threads, nil
}

func Thread(board string, op int64) ([]Post, error) {
	sq3, err := validate(board)
	if err != nil {
		return nil, err
	}
	defer sq3.Close()

	stmt, err := sq3.Prepare(fmt.Sprintf("SELECT * FROM %s_posts "+
		"WHERE op = %d "+
		"ORDER BY id", board, op))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	posts := make([]Post, 0, 64)
	for {
		if ok, err := stmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		var post Post
		err = stmt.Scan(
			&post.Id,
			&post.Op,
			&post.Ip,
			&post.Date,
			&post.Name,
			&post.Email,
			&post.Subject,
			&post.Body)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		return nil, OpNotFound(op)
	}
	return posts, nil
}

func Op(board string, id int64) (int64, error) {
	sq3, err := validate(board)
	if err != nil {
		return 0, err
	}
	defer sq3.Close()

	stmt, err := sq3.Prepare(fmt.Sprintf("SELECT op FROM %s_posts WHERE id = %d", board, id))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	if ok, err := stmt.Step(); err != nil {
		return 0, err
	} else if !ok {
		return 0, IdNotFound(id)
	}
	var op int64
	err = stmt.Scan(&op)
	return op, err
}

func Reply(board string, op int64, ip string, name, email, subject, body string) error {
	sq3, err := validate(board)
	if err != nil {
		return err
	}
	defer sq3.Close()

	if op > 0 {
		stmt, err := sq3.Prepare(fmt.Sprintf("SELECT op FROM %s_posts "+
			"WHERE op = %d "+
			"LIMIT 1", board, op))
		if err != nil {
			return err
		}
		defer stmt.Close()
		if ok, err := stmt.Step(); err != nil {
			return err
		} else if !ok {
			return OpNotFound(op)
		}
	}

	err = sq3.Exec(fmt.Sprintf("INSERT INTO %s_posts VALUES("+
		"NULL, %d, '%s', datetime('now'), ?, ?, ?, ?)", board, op, ip),
		name, email, subject, body)
	if err != nil {
		return err
	}
	if op <= 0 {
		return sq3.Exec(fmt.Sprintf("UPDATE %s_posts "+
			"SET op = id "+
			"WHERE op = %d", board, op))
	}
	return nil
}
