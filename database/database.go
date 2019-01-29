package database

import (
	"fmt"
	"html/template"

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
			"AFTER INSERT ON %s_posts WHEN NEW.op = -1 BEGIN "+
			"INSERT INTO %s_ops VALUES("+
			"(SELECT id FROM %s_posts WHERE op = -1 LIMIT 1),"+
			"datetime('now'));"+
			"UPDATE %s_posts SET op = id WHERE op = -1;"+
			"END",
			b.Name, b.Name, b.Name, b.Name, b.Name))
		if err != nil {
			return err
		}
		err = conn.Exec(fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %s_insert_post "+
			"AFTER INSERT ON %s_posts WHEN NEW.op != -1 AND "+
			"NEW.email IS NOT 'sage' BEGIN "+
			"UPDATE %s_ops SET bumped = datetime('now') WHERE id = NEW.op;"+
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
	return fmt.Sprintf("Board not found: %s", string(b))
}

type OpNotFound int64

func (op OpNotFound) Error() string {
	return fmt.Sprintf("Op not found: %d", int64(op))
}

type IdNotFound int64

func (id IdNotFound) Error() string {
	return fmt.Sprintf("Post not found: %d", int64(id))
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

func scanPost(postStmt, imgStmt *sqlite3.Stmt) (*Post, error) {
	post := new(Post)
	comment := ""
	post.Image = new(ImageAttr)
	err := postStmt.Scan(
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
	if post.Name == "" {
		post.Name = "Anonymous"
	}
	post.Comment = template.HTML(comment)
	return post, nil
}

type Preview struct {
	Op      *Post
	Replies []*Post
}

func GetBoard(board string) ([]Preview, error) {
	conn, exit, err := enter(board)
	if err != nil {
		return nil, err
	}
	defer exit(conn)

	/* i assume there's a better way that reduces the number of trips to
	 * the db but i know basically nothing about sql (: */
	opStmt, err := conn.Prepare(fmt.Sprintf("SELECT id FROM %s_ops "+
		"ORDER BY bumped DESC LIMIT 10",
		board))
	if err != nil {
		return nil, err
	}
	defer opStmt.Close()
	postStmt, err := conn.Prepare(fmt.Sprintf("SELECT * FROM %s_posts "+
		"WHERE id = ? UNION "+
		"SELECT * FROM "+
		"(SELECT * FROM %s_posts WHERE op = ? "+
		"ORDER BY id DESC LIMIT 3)", board, board))
	if err != nil {
		return nil, err
	}
	defer postStmt.Close()
	imgStmt, err := conn.Prepare("SELECT size, width, height FROM images WHERE uri = ?")
	if err != nil {
		return nil, err
	}
	defer imgStmt.Close()

	threads := make([]Preview, 0, 10)
	for {
		if ok, err := opStmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		var op int64
		err = opStmt.Scan(&op)
		if err != nil {
			return nil, err
		}
		if err = postStmt.Bind(op, op); err != nil {
			return nil, err
		}
		var prv Preview
		if ok, err := postStmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			continue // ???
		}
		post, err := scanPost(postStmt, imgStmt)
		if err != nil {
			return nil, err
		}
		prv.Op = post
		for {
			if ok, err := postStmt.Step(); err != nil {
				return nil, err
			} else if !ok {
				break
			}

			post, err := scanPost(postStmt, imgStmt)
			if err != nil {
				return nil, err
			}
			prv.Replies = append(prv.Replies, post)
		}
		if err = postStmt.Reset(); err != nil {
			return nil, err
		}
		threads = append(threads, prv)
	}
	return threads, nil
}

func GetThread(board string, op int64) ([]*Post, error) {
	conn, exit, err := enter(board)
	if err != nil {
		return nil, err
	}
	defer exit(conn)

	postStmt, err := conn.Prepare(fmt.Sprintf("SELECT * FROM %s_posts "+
		"WHERE op = %d "+
		"ORDER BY id",
		board, op))
	if err != nil {
		return nil, err
	}
	defer postStmt.Close()
	imgStmt, err := conn.Prepare("SELECT size, width, height FROM images WHERE uri = ?")
	if err != nil {
		return nil, err
	}
	defer imgStmt.Close()

	posts := make([]*Post, 0, 64)
	for {
		if ok, err := postStmt.Step(); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		post, err := scanPost(postStmt, imgStmt)
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

// the use of interface{} here is due to go-sqlite-lite, not go. i (stupidly)
// want sql nulls and go-sqlite-lite (sanely) treats the empty string as
// non-null and byte slices as blobs. unfortunately writing a custom bind
// function isn't an option either as the package doesn't export the
// sqlite_bind_* functions and i don't feel like forking the package
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

	if op > 0 {
		stmt, err := conn.Prepare(fmt.Sprintf("SELECT NULL FROM %s_posts "+
			"WHERE op = %d "+
			"LIMIT 1",
			board, op))
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
		if req.Image == nil {
			return conn.Exec(fmt.Sprintf("INSERT INTO %s_posts VALUES("+
				"NULL, %d, '%s', datetime('now'), "+
				"?, ?, ?, ?, "+
				"?, ?, ?)",
				board, op, ip),
				req.Name, req.Email, req.Subject, req.Comment,
				nil, nil, nil)
		}
		uri, err := submitImage(conn, req.Image)
		if err != nil {
			return err
		}
		return conn.Exec(fmt.Sprintf("INSERT INTO %s_posts VALUES("+
			"NULL, %d, '%s', datetime('now'), "+
			"?, ?, ?, ?, "+
			"?, ?, ?)",
			board, op, ip),
			req.Name, req.Email, req.Subject, req.Comment,
			req.ImageName, req.ImageAlt, uri)
	})
}
