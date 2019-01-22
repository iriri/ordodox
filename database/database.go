package database

import (
	"fmt"
	"html/template"
	"strconv"

	"github.com/bvinc/go-sqlite-lite/sqlite3"

	"ordodox/config"
)

const OpId = -1

const nconns = 8

var path string
var conns = make(chan *sqlite3.Conn, nconns)
var Boards = make(map[string]string)

func init_(conn *sqlite3.Conn, boards []config.Board) error {
	err := conn.Exec("CREATE TABLE IF NOT EXISTS images(" +
		"uri TEXT PRIMARY KEY," +
		"data BLOB NOT NULL," +
		"thumb BLOB NOT NULL," +
		"size INT," +
		"width INT," +
		"height INT) WITHOUT ROWID")
	if err != nil {
		return err
	}
	for _, b := range boards {
		Boards[b.Name] = b.Title
		// i don't care about startup time so these Sprintfs can stay
		err = conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_posts("+
			"id INTEGER PRIMARY KEY,"+
			"op INT NOT NULL,"+
			"ip TEXT NOT NULL,"+
			"date DATE NOT NULL,"+
			"name TEXT,"+
			"email TEXT,"+
			"subject TEXT,"+
			"comment TEXT,"+
			"imagename TEXT,"+
			"imagealt TEXT,"+
			"imageuri TEXT REFERENCES images(uri) ON DELETE SET NULL)",
			b.Name))
		if err != nil {
			return err
		}
		err = conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_ops("+
			"id INT PRIMARY KEY REFERENCES %s_posts(id) ON DELETE CASCADE,"+
			"bumped DATE NOT NULL) WITHOUT ROWID",
			b.Name, b.Name))
		if err != nil {
			return err
		}
		err = conn.Exec(fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %s_insert_op "+
			"AFTER INSERT ON %s_posts WHEN NEW.op=-1 BEGIN "+
			"INSERT INTO %s_ops VALUES("+
			"(SELECT id FROM %s_posts WHERE op=-1 LIMIT 1),"+
			"datetime('now'));"+
			"UPDATE %s_posts SET op=id WHERE op=-1;"+
			"END",
			b.Name, b.Name, b.Name, b.Name, b.Name))
		if err != nil {
			return err
		}
		err = conn.Exec(fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %s_insert_post "+
			"AFTER INSERT ON %s_posts WHEN NEW.op!=-1 AND "+
			"NEW.email IS NOT 'sage' BEGIN "+
			"UPDATE %s_ops SET bumped=datetime('now') WHERE id=NEW.op;"+
			"END",
			b.Name, b.Name, b.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

func Init(opt *config.Opt) error {
	path = opt.Db
	initCaches()

	for i := 1; i < nconns; i++ {
		conn, err := sqlite3.Open(path)
		if err != nil {
			return err
		}
		conns <- conn
	}
	conn, err := sqlite3.Open(path)
	if err != nil {
		return err
	}
	defer func() { conns <- conn }()
	return conn.WithTxImmediate(func() error {
		return init_(conn, opt.Boards)
	})
}

type BoardNotFound string

func (b BoardNotFound) Error() string {
	return "Board not found: " + string(b)
}

type OpNotFound int64

func (op OpNotFound) Error() string {
	return "Op not found: " + strconv.FormatInt(int64(op), 10)
}

type IdNotFound int64

func (id IdNotFound) Error() string {
	return "Post not found: " + strconv.FormatInt(int64(id), 10)
}

func getConn() (*sqlite3.Conn, func(*sqlite3.Conn), error) {
	select {
	case conn := <-conns:
		return conn, func(conn *sqlite3.Conn) { conns <- conn }, nil
	default:
		conn, err := sqlite3.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return conn, func(conn *sqlite3.Conn) { conn.Close() }, nil
	}
}

func enter(board string) (*sqlite3.Conn, func(*sqlite3.Conn), error) {
	if _, ok := Boards[board]; !ok {
		// in theory an sql injection might still be possible if the
		// config has an exceptionally stupid board name like "; DROP
		// TABLES *; --" or something idk
		return nil, nil, BoardNotFound(board)
	}
	return getConn()
}

type ImageAttr struct {
	Name   string
	Alt    string
	Uri    string
	Size   int64
	Width  int64
	Height int64
}

type Post struct {
	Id      int64
	Op      int64
	Date    string
	Name    string
	Email   string
	Subject string
	Comment template.HTML
	Image   *ImageAttr // pointer to avoid marshalling when nil... should consider alternatives
}

func GetBoard(board string) ([][]*Post, error) {
	conn, exit, err := enter(board)
	if err != nil {
		return nil, err
	}
	defer exit(conn)

	stmt, err := conn.Prepare("SELECT id FROM " + board + "_ops " +
		"ORDER BY bumped DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	threads := make([][]*Post, 0, 10)
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
		threads = append(threads, []*Post{&Post{Id: op}})
	}
	return threads, nil
}

func GetThread(board string, op int64) ([]*Post, error) {
	conn, exit, err := enter(board)
	if err != nil {
		return nil, err
	}
	defer exit(conn)

	postStmt, err := conn.Prepare("SELECT * FROM " + board + "_posts " +
		"WHERE op=" + strconv.FormatInt(op, 10) + " ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer postStmt.Close()
	imgStmt, err := conn.Prepare("SELECT size, width, height FROM images WHERE uri=?")
	if err != nil {
		return nil, err
	}
	posts := make([]*Post, 0, 64)
	for {
		if ok, err := postStmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		post := new(Post)
		comment := ""
		post.Image = new(ImageAttr)
		err = postStmt.Scan(
			&post.Id,
			&post.Op,
			nil,
			&post.Date,
			&post.Name,
			&post.Email,
			&post.Subject,
			&comment,
			&post.Image.Name,
			&post.Image.Alt,
			&post.Image.Uri)
		if err != nil {
			return nil, err
		}
		if post.Image.Uri == "" {
			post.Image = nil
		} else {
			if err = imgStmt.Bind(post.Image.Uri); err != nil {
				return nil, err
			}
			if ok, err := imgStmt.Step(); err != nil {
				return nil, err
			} else if !ok {
				// almost impossible due to the sql foreign key trigger
				post.Image = nil
				goto imgDeleted
			}
			err = imgStmt.Scan(&post.Image.Size, &post.Image.Width, &post.Image.Height)
			if err != nil {
				return nil, err
			}
			if err = imgStmt.Reset(); err != nil {
				return nil, err
			}
		}
	imgDeleted:
		post.Comment = template.HTML(comment)
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		return nil, OpNotFound(op)
	}
	return posts, nil
}

// sadly, i want sql nulls and i don't feel like forking go-sqlite-lite which
// treats the empty string as non-null and byte slices as blobs. this is the
// sane thing to do, of course, but unfortunately the package doesn't export
// the sqlite_bind_* functions meaning that the user can't just write their own
// custom bind function
type Request struct {
	Name      interface{}
	Email     interface{}
	Subject   interface{}
	Comment   interface{}
	Image     []byte
	ImageName string
	ImageAlt  interface{}
}

func Submit(board string, op int64, ip string, req *Request) error {
	conn, exit, err := enter(board)
	if err != nil {
		return err
	}
	defer exit(conn)

	op_ := strconv.FormatInt(op, 10)
	if op > 0 {
		stmt, err := conn.Prepare("SELECT NULL FROM " + board + "_posts " +
			"WHERE op=" + op_ + " LIMIT 1")
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

	req.Comment, err = parse(conn, board, req.Comment, op)
	if err != nil {
		return err
	}
	return conn.WithTx(func() error {
		query := "INSERT INTO " + board + "_posts VALUES(" +
			"NULL," + op_ + ",'" + ip + "',datetime('now')," +
			"?,?,?,?,?,?,?)"
		params := make([]interface{}, 7)
		params[0] = req.Name
		params[1] = req.Email
		params[2] = req.Subject
		params[3] = req.Comment
		if req.Image == nil {
			return conn.Exec(query, params...)
		}

		uri, err := submitImage(conn, req.Image)
		if err != nil {
			return err
		}
		params[4] = req.ImageName
		params[5] = req.ImageAlt
		params[6] = uri
		return conn.Exec(query, params...)
	})
}
