package database

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"unsafe"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"golang.org/x/crypto/blowfish"
)

// tripcode generation is based on bcrypt, originally designed by Provos and
// Mazières, and this implementation borrows VERY heavily from
// golang.org/x/crypto/bcrypt, which is BSD licensed and copyright 2011, The Go
// Authors. unfortunately, i need a slightly different api
var magicCipherData = []byte{
	0x4f, 0x72, 0x70, 0x68,
	0x65, 0x61, 0x6e, 0x42,
	0x65, 0x68, 0x6f, 0x6c,
	0x64, 0x65, 0x72, 0x53,
	0x63, 0x72, 0x79, 0x44,
	0x6f, 0x75, 0x62, 0x74,
}

func getDbSalt(conn *sqlite3.Conn, name string) ([]byte, []byte, error) {
	stmt, err := conn.Prepare("SELECT salt, hash FROM trips WHERE name = ?")
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()
	var salt, hash []byte
	if err = stmt.Bind(name); err != nil {
		return nil, nil, err
	}
	if ok, err := stmt.Step(); err != nil {
		return nil, nil, err
	} else if ok {
		err = stmt.Scan(&salt, &hash)
		return salt, hash, err
	}
	return nil, nil, nil
}

func getSalt(conn *sqlite3.Conn, name string) ([]byte, []byte, error) {
	salt, hash, err := getDbSalt(conn, name)
	if err != nil {
		return nil, nil, err
	}
	if salt != nil {
		return salt, hash, err
	}

	salt = make([]byte, 16)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}
	if err = conn.Exec("INSERT INTO trips VALUES(?, ?, NULL)", name, salt); err == nil {
		return salt, nil, nil
	} else if err := err.(*sqlite3.Error); err.Code() != sqlite3.CONSTRAINT_PRIMARYKEY {
		return nil, nil, err
	}
	return getDbSalt(conn, name)
}

func expensiveBlowfishSetup(key string, salt []byte) (*blowfish.Cipher, error) {
	key_ := make([]byte, len(key))
	copy(key_, key)
	salt_ := make([]byte, len(salt))
	copy(salt_, salt)
	c, err := blowfish.NewSaltedCipher(key_, salt_)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 1<<12; i++ {
		blowfish.ExpandKey(key_, c)
		blowfish.ExpandKey(salt_, c)
	}
	return c, nil
}

func bcrypt(key string, salt []byte) ([]byte, error) {
	data := make([]byte, len(magicCipherData))
	copy(data, magicCipherData)
	c, err := expensiveBlowfishSetup(key, salt)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 24; i += 8 {
		for j := 0; j < 64; j++ {
			c.Encrypt(data[i:i+8], data[i:i+8])
		}
	}
	return data, nil
}

const alphabet = "./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var bcEncoding = base64.NewEncoding(alphabet)

func encode(hash []byte, match bool) string {
	enc := make([]byte, 13)
	if match {
		enc[0] = '!'
	} else {
		enc[0] = '?'
	}
	bcEncoding.Encode(enc[1:], hash[15:])
	return *(*string)(unsafe.Pointer(&enc))
}

var alreadyUpdated = errors.New("Row was already updated")

func tripcode(conn *sqlite3.Conn, name string) (string, interface{}, error) {
	parts := strings.SplitN(name, "#", 2)
	if len(parts) == 1 {
		return parts[0], nil, nil
	}
	salt, hash, err := getSalt(conn, parts[0])
	if err != nil {
		return "", nil, err
	}
	hashed, err := bcrypt(parts[1], salt)
	if err != nil {
		return "", nil, err
	}
	if hash != nil {
		return parts[0], encode(hashed, subtle.ConstantTimeCompare(hash, hashed) == 1), nil
	}

	err = conn.WithTx(func() error {
		err := conn.Exec("UPDATE trips SET hash = ? "+
			"WHERE name = ? AND hash IS NULL",
			hashed, parts[0])
		if err != nil {
			return err
		}
		stmt, err := conn.Prepare("SELECT changes()")
		if err != nil {
			return err
		}
		if ok, err := stmt.Step(); err != nil {
			return err
		} else if !ok {
			panic("what")
		}

		var changes int64
		if err = stmt.Scan(&changes); err != nil {
			return err
		}
		switch changes {
		case 0:
			return alreadyUpdated
		case 1:
			return nil
		default:
			panic("transactions aren't atomic?")
		}
	})
	if err == nil {
		return parts[0], encode(hashed, true), nil
	} else if err != alreadyUpdated {
		return "", nil, err
	}
	if _, hash, err = getDbSalt(conn, parts[0]); err != nil {
		return "", nil, err
	}
	return parts[0], encode(hashed, subtle.ConstantTimeCompare(hash, hashed) == 1), nil
}
