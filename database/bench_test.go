// BenchmarkUnsafe-4    	20000000	        55.7 ns/op
// BenchmarkCopy-4      	10000000	       223 ns/op
// BenchmarkConcat-4    	 5000000	       284 ns/op
// BenchmarkBuilder-4   	 5000000	       299 ns/op
// BenchmarkJoin-4      	 3000000	       597 ns/op
// BenchmarkSprintf-4   	 1000000	      1016 ns/op
//
// time to banish Sprintf from any paths that matter
package database

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

var board = "a"

func BenchmarkUnsafe(b *testing.B) {
	bs := make([]byte, 256)
	for i := 0; i < b.N; i++ {
		n := copy(bs, "CREATE TRIGGER IF NOT EXISTS ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_insert_op AFTER INSERT ON ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts WHEN NEW.op = -1 BEGIN INSERT INTO ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_ops VALUES((SELECT id FROM ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts WHERE op = -1 LIMIT 1),datetime('now'));UPDATE ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts SET op = id WHERE op = -1;END")
		_ = *(*string)(unsafe.Pointer(&bs))
	}
}

func BenchmarkCopy(b *testing.B) {
	bs := make([]byte, 256)
	for i := 0; i < b.N; i++ {
		n := copy(bs, "CREATE TRIGGER IF NOT EXISTS ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_insert_op AFTER INSERT ON ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts WHEN NEW.op = -1 BEGIN INSERT INTO ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_ops VALUES((SELECT id FROM ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts WHERE op = -1 LIMIT 1),datetime('now'));UPDATE ")
		n += copy(bs[n:], board)
		n += copy(bs[n:], "_posts SET op = id WHERE op = -1;END")
		_ = string(bs)
	}
}

func BenchmarkConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "CREATE TRIGGER IF NOT EXISTS " + board + "_insert_op " +
			"AFTER INSERT ON " + board + "_posts WHEN NEW.op = -1 BEGIN " +
			"INSERT INTO " + board + "_ops VALUES(" +
			"(SELECT id FROM " + board + "_posts WHERE op = -1 LIMIT 1)," +
			"datetime('now'));" +
			"UPDATE " + board + "_posts SET op = id WHERE op = -1;" +
			"END"
	}
}

func BenchmarkBuilder(b *testing.B) {
	sb := new(strings.Builder)
	for i := 0; i < b.N; i++ {
		sb.WriteString("CREATE TRIGGER IF NOT EXISTS ")
		sb.WriteString(board)
		sb.WriteString("_insert_op AFTER INSERT ON ")
		sb.WriteString(board)
		sb.WriteString("_posts WHEN NEW.op = -1 BEGIN INSERT INTO ")
		sb.WriteString(board)
		sb.WriteString("_ops VALUES((SELECT id FROM ")
		sb.WriteString(board)
		sb.WriteString("_posts WHERE op = -1 LIMIT 1),datetime('now'));UPDATE ")
		sb.WriteString(board)
		sb.WriteString("_posts SET op = id WHERE op = -1;END")
		_ = sb.String()
	}
}

func BenchmarkJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strings.Join([]string{"CREATE TRIGGER IF NOT EXISTS ", board, "_insert_op ",
			"AFTER INSERT ON ", board, "_posts WHEN NEW.op = -1 BEGIN ",
			"INSERT INTO ", board, "_ops VALUES(",
			"(SELECT id FROM ", board, "_posts WHERE op = -1 LIMIT 1),",
			"datetime('now'));",
			"UPDATE ", board, "_posts SET op = id WHERE op = -1;",
			"END"}, "")
	}
}

func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %s_insert_op "+
			"AFTER INSERT ON %s_posts WHEN NEW.op = -1 BEGIN "+
			"INSERT INTO %s_ops VALUES("+
			"(SELECT id FROM %s_posts WHERE op = -1 LIMIT 1),"+
			"datetime('now'));"+
			"UPDATE %s_posts SET op = id WHERE op = -1;"+
			"END",
			board, board, board, board, board)
	}
}
