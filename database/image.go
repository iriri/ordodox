package database

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"strings"

	"github.com/bamiaux/rez"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

func normalizeSuffix(typ string) string {
	if typ == "jpeg" {
		return "jpg"
	}
	return typ
}

func makeThumb(input image.Image, size image.Point) ([]byte, error) {
	var x, y, factor float64
	var x_, y_ int
	x, y = float64(size.X), float64(size.Y)
	if x > y {
		factor = 160 / x
	} else {
		factor = 160 / y
	}
	if factor > 1 {
		x_, y_ = size.X, size.Y
	} else {
		x_, y_ = int((x*factor)+0.5), int((y*factor)+0.5)
	}
	var output image.Image
	switch input.(type) {
	case *image.RGBA:
		output = image.NewRGBA(image.Rect(0, 0, x_, y_))
	case *image.YCbCr:
		output = image.NewYCbCr(image.Rect(0, 0, x_, y_), image.YCbCrSubsampleRatio420)
	default:
		return nil, errors.New("greyscale or something? idk deal with this later")
	}

	err := rez.Convert(output, input, rez.NewBilinearFilter())
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, output, &jpeg.Options{90})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func submitImage(conn *sqlite3.Conn, input []byte) (string, error) {
	r := bytes.NewReader(input)
	img, typ, err := image.Decode(r)
	if err != nil {
		return "", err
	}
	sum := sha512.Sum512_256(input)
	name := make([]byte, 65, 68)
	hex.Encode(name[0:], sum[:])
	name[64] = '.'
	name = append(name, normalizeSuffix(typ)...)
	name_ := string(name)
	checkStmt, err := conn.Prepare(fmt.Sprintf(
		"SELECT NULL FROM images WHERE uri = '%s'",
		name_))
	if err != nil {
		return "", err
	}
	defer checkStmt.Close()
	if ok, err := checkStmt.Step(); err != nil {
		return "", err
	} else if ok {
		return name_, nil
	}

	size := img.Bounds().Size()
	thumb, err := makeThumb(img, size)
	if err != nil {
		return "", err
	}
	insertStmt, err := conn.Prepare(fmt.Sprintf("INSERT INTO images "+
		"VALUES('%s', ?, ?, %d, %d, %d)",
		name_, len(input)/1024, size.X, size.Y))
	if err != nil {
		return "", err
	}
	defer insertStmt.Close()
	err = insertStmt.Bind(input, thumb)
	if err != nil {
		return "", err
	}
	if _, err := insertStmt.Step(); err != nil {
		return "", err
	}
	return name_, nil
}

type ImageNotFound string

func (uri ImageNotFound) Error() string {
	return fmt.Sprintf("Image not found: %s", uri)
}

func getImage(uri, kind string) ([]byte, error) {
	conn, exit, err := getConn()
	if err != nil {
		return nil, err
	}
	defer exit(conn)

	stmt, err := conn.Prepare(fmt.Sprintf("SELECT %s FROM images WHERE uri = ?", kind))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Bind(uri)
	if err != nil {
		return nil, err
	}
	if ok, err := stmt.Step(); err != nil {
		return nil, err
	} else if !ok {
		return nil, ImageNotFound(uri)
	}
	return stmt.ColumnBlob(0)
}

func GetImage(uri string) ([]byte, error) {
	if len(uri) != 68 {
		return nil, ImageNotFound(uri)
	}
	return getImage(uri, "data")
}

func GetThumb(uri string) ([]byte, error) {
	if len(uri) != 78 || !strings.HasSuffix(uri, ".thumb.jpg") {
		return nil, ImageNotFound(uri)
	}
	return getImage(uri[:68], "thumb")
}
