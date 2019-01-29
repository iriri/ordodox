package database

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"unicode"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

// this probably wasn't the right approach...
type tokType int

const (
	str tokType = iota
	bs
	gt
	gtgt
	grv
	grvgrvgrv
	eol
	eof
)

type token struct {
	typ tokType
	val []rune
}

type lexer struct {
	input []rune
	idx   int
}

func (l *lexer) gt() token {
	if l.input[l.idx+1] != '>' {
		return token{typ: gt}
	}

	l.idx += 2
	j := l.idx
	for ; l.idx < len(l.input) && unicode.IsDigit(l.input[l.idx]); l.idx++ {
	}
	var tok token
	if j < l.idx {
		tok = token{typ: gtgt, val: l.input[j:l.idx]}
	} else {
		tok = token{typ: gtgt}
	}
	if l.idx < len(l.input) {
		l.idx--
	}
	return tok
}

func (l *lexer) grv() token {
	if l.input[l.idx+1] != '`' {
		return token{typ: grv}
	}
	if l.idx+2 < len(l.input) && l.input[l.idx+2] == '`' {
		l.idx += 2
		return token{typ: grvgrvgrv}
	}
	l.idx++
	return token{typ: grv}
}

func (l *lexer) str() token {
	j := l.idx
	for ; l.idx < len(l.input); l.idx++ {
		if l.input[l.idx] == '\\' ||
			l.input[l.idx] == '>' ||
			l.input[l.idx] == '`' ||
			l.input[l.idx] == '\n' {
			break
		}
	}
	tok := token{typ: str, val: l.input[j:l.idx]}
	if l.idx < len(l.input) {
		l.idx--
	}
	return tok
}

func (l *lexer) token() token {
	var tok token
	switch l.input[l.idx] {
	case '\\':
		tok = token{typ: bs}
	case '>':
		tok = l.gt()
	case '`':
		tok = l.grv()
	case '\n':
		tok = token{typ: eol}
	default:
		tok = l.str()
	}
	l.idx++
	return tok
}

func (l *lexer) lex() []token {
	toks := make([]token, 0, 8)
	for l.idx < len(l.input)-1 {
		toks = append(toks, l.token())
	}
	if l.idx < len(l.input) {
		switch l.input[l.idx] {
		case '>':
			toks = append(toks, token{typ: gt})
		case '`':
			toks = append(toks, token{typ: grv})
		default:
			toks = append(toks, token{typ: str, val: l.input[l.idx : l.idx+1]})
		}
	}
	return append(toks, token{typ: eof})
}

type node interface {
	nodeMarker()
}

type text string

func (text) nodeMarker() {}

type greenText []node

func (greenText) nodeMarker() {}

type idRef string

func (idRef) nodeMarker() {}

type code string

func (code) nodeMarker() {}

type blockCode string

func (blockCode) nodeMarker() {}

type term struct{}

func (term) nodeMarker() {}

// wait this is a fucking disaster fuck
type parser struct {
	toks []token
	idx  int
}

func (p *parser) shift() token {
	tok := p.toks[p.idx]
	p.idx++
	return tok
}

func (p *parser) escaped() node {
	switch tok := p.shift(); tok.typ {
	case str:
		return text(tok.val)
	case bs:
		return text("\\")
	case gt:
		return text(">")
	case gtgt:
		return text(">>")
	case grv:
		return text("`")
	case grvgrvgrv:
		return text("```")
	case eol:
		return text(" ")
	case eof:
		return term{}
	}
	panic("unreachable")
}

func (p *parser) greenText() node {
	sub := make([]node, 0, 8)
	for p.idx < len(p.toks) {
		switch tok := p.shift(); tok.typ {
		case str:
			sub = append(sub, text(tok.val))
		case bs:
			switch t := p.escaped().(type) {
			case text:
				sub = append(sub, t)
			case term:
				return greenText(sub)
			}
		case gt:
			sub = append(sub, text(">"))
		case gtgt:
			sub = append(sub, idRef(tok.val))
		case grv:
			sub = append(sub, p.code())
		case grvgrvgrv:
			p.idx--
			fallthrough
		case eol, eof:
			return greenText(sub)
		}
	}
	return greenText(sub)
}

func (p *parser) code() node {
	sb := new(strings.Builder)
	for p.idx < len(p.toks) {
		switch tok := p.shift(); tok.typ {
		case str:
			for _, r := range tok.val {
				sb.WriteRune(r)
			}
		case bs:
			switch t := p.escaped().(type) {
			case text:
				sb.WriteString(string(t))
			case term:
				return code(sb.String())
			}
		case gt:
			sb.WriteString(">")
		case gtgt:
			sb.WriteString(">>")
		case eol:
			sb.WriteString(" ")
		case grv, grvgrvgrv, eof:
			return code(sb.String())
		}
	}
	return code(sb.String())
}

func (p *parser) blockCode() node {
	sb := new(strings.Builder)
	for p.idx < len(p.toks) {
		switch tok := p.shift(); tok.typ {
		case str:
			for _, r := range tok.val {
				sb.WriteRune(r)
			}
		case bs:
			switch t := p.escaped().(type) {
			case text:
				sb.WriteString(string(t))
			case term:
				return blockCode(sb.String())
			}
		case gt:
			sb.WriteString(">")
		case gtgt:
			sb.WriteString(">>")
		case grv:
			sb.WriteString("`")
		case eol:
			sb.WriteString("\n")
		case grvgrvgrv, eof:
			return blockCode(sb.String())
		}
	}
	return blockCode(sb.String())
}

func (p *parser) node() node {
	switch tok := p.shift(); tok.typ {
	case str:
		return text(tok.val)
	case bs:
		return p.escaped()
	case gt:
		return p.greenText()
	case gtgt:
		return idRef(tok.val)
	case grv:
		return p.code()
	case grvgrvgrv:
		return p.blockCode()
	case eol:
		return term{}
	case eof:
		// return term{}
	}
	panic("unreachable")
}

func foldTerms(ast []node) []node {
	i := len(ast) - 1
	for ; i >= 0; i-- {
		if _, ok := ast[i].(term); !ok {
			break
		}
	}
	return ast[:i+1]
}

func (p *parser) parse() []node {
	ast := make([]node, 0, 8)
	for p.idx < len(p.toks)-1 {
		ast = append(ast, p.node())
	}
	return foldTerms(ast)
}

func emitIdRef(conn *sqlite3.Conn, board string, op int64, sb *strings.Builder, ref string) error {
	id, err := strconv.ParseInt(ref, 10, 64)
	if err != nil {
		sb.WriteString("&gt;&gt;")
		sb.WriteString(html.EscapeString(ref))
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
		sb.WriteString("&gt;&gt;")
		sb.WriteString(html.EscapeString(ref))
		return nil
	}
	var op_ int64
	err = stmt.Scan(&op_)
	if err != nil {
		return err
	}

	sb.WriteString("<a href=\"")
	if op != op_ {
		sb.WriteByte('/')
		sb.WriteString(board)
		sb.WriteByte('/')
		sb.WriteString(strconv.FormatInt(op_, 10))
	}
	sb.WriteByte('#')
	sb.WriteString(ref)
	sb.WriteString("\">&gt;&gt;")
	sb.WriteString(ref)
	sb.WriteString("</a>")
	return nil
}

func emit(conn *sqlite3.Conn, board string, op int64, ast []node) (string, error) {
	sb := new(strings.Builder)
	for _, n := range ast {
		switch n := n.(type) {
		case text:
			sb.WriteString(html.EscapeString(string(n)))
		case greenText:
			sb.WriteString("<span class=\"gt\">&gt;")
			if sub, err := emit(conn, board, op, n); err != nil {
				return "", err
			} else {
				sb.WriteString(sub)
			}
			sb.WriteString("</span><br>")
		case idRef:
			if err := emitIdRef(conn, board, op, sb, string(n)); err != nil {
				return "", err
			}
		case code:
			if len(n) > 0 {
				sb.WriteString("<pre>")
				sb.WriteString(html.EscapeString(string(n)))
				sb.WriteString("</pre>")
			}
		case blockCode:
			if len(n) > 0 {
				sb.WriteString("<div class=\"block\"><pre>")
				sb.WriteString(html.EscapeString(string(n)))
				sb.WriteString("</pre></div>")
			}
		case term:
			sb.WriteString("<br>")
		}
	}
	return sb.String(), nil
}

func parse(conn *sqlite3.Conn, board string, comm interface{}, op int64) (interface{}, error) {
	comm_, ok := comm.(string)
	if !ok {
		if comm != nil {
			panic("wait what the fuck?")
		}
		return nil, nil
	}
	l := &lexer{input: []rune(comm_)}
	p := &parser{toks: l.lex()}
	return emit(conn, board, op, p.parse())
}
