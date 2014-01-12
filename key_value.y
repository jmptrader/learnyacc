%{
package main

import (
	"fmt"
	"bufio"
	"os"
	"io"
	"log"
	"unicode"
	"bytes"
)
%}

%union{
	tok int
	val interface{}
	pair struct{key, val interface{}}
	pairs map[interface{}]interface{}
}

%token <val> IDENT

%type <pair> pair
%type <pairs> pairs

%%

goal:
	'{' pairs '}'
	{
		yylex.(*lex).m = $2
	}

pairs:
	pair
	{
		$$ = map[interface{}]interface{}{$1.key: $1.val}
	}
|   pairs '|' pair
	{
		$$[$3.key] = $3.val
	}

pair:
	IDENT '=' IDENT
	{
		$$.key, $$.val = $1, $3
	}
|   IDENT '=' '{' pairs '}'
	{
		$$.key, $$.val = $1, $4
	}


%%

type lex struct {
	reader *bufio.Reader
	next rune
	m map[interface{}]interface{}
}

var terminals map[rune]bool

func (l *lex) Lex(lval *yySymType) int {
	next := l.next
	var cur rune
	eat_id := false
	var ident *bytes.Buffer
	for {
		cur = next
		next, _, err := l.reader.ReadRune()
		if err != nil {
			if err != io.EOF {
				l.Error(err.Error())
			}
		}

		if unicode.IsSpace(cur) {
			continue
		}

		if terminals[cur] {
			lval.val = ""
			l.next = next
			return int(cur)
		}
		if unicode.IsLetter(cur) {
			if !eat_id {
				eat_id = true
				ident = new(bytes.Buffer)
			}
			_, err := ident.WriteRune(cur)
			if err != nil {
				l.Error(err.Error())
			}
			if !unicode.IsLetter(next) {
				lval.val = ident.String()
				l.next = next
				return IDENT
			}
		}
	}
}

func (l *lex) Error(e string) {
	log.Fatal(e)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	terminals = map[rune]bool{
		'{': true,
		'=': true,
		'|': true,
	}
	l := &lex{
		reader,
		' ',
		map[interface{}]interface{}{},
	}
	yyParse(l)
	fmt.Println(l.m)
}
