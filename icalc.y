%{
package main

import (
	"fmt"
	"os"
	"math"
	"bufio"
	"io"
	"unicode/utf8"
	"strconv"
)

%}

%union{
	val float64
}

%token <val> NUM
%token NEG

%left '-' '+'
%left '*' '/'
%left NEG		// negation--unary minus
%right '^'			// exponentiation

%type <val> exp

%%

input:
	/* empty */
	| input line


line:
	'\n'
	| exp '\n'		{ fmt.Printf("HEYYA it's a: %.10g\n", $1) }


exp:
	NUM					{ $$ = $1           }
	| exp '+' exp		{ $$ = $1 + $3      }
	| exp '-' exp		{ $$ = $1 - $3      }
	| exp '*' exp		{ $$ = $1 * $3      }
	| exp '/' exp		{ $$ = $1 / $3      }
	| '-' exp %prec NEG	{ $$ = -$2          }  /* Unary minus    */
	| exp '^' exp	{ $$ = math.Pow($1, $3) }  /* Exponentiation */
	| '(' exp ')'		{ $$ = $2           }

%%

var peekrune rune = ' '
var line string = ""
var sym string = ""
var nerrors int = 0
var lineno int = 1
var linep int = 0
var reader *bufio.Reader

var numberStartRunes = map[rune]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
	'.': true,
}

var numberRunes = make(map[rune]bool)
var numberMiddleOtherRunes = map[rune]bool{
	'e': true,
	'E': true,
	'x': true,
	'X': true,
	'-': true,
	'+': true,
}

var spaceRunes = map[rune]bool{
	' ': true,
	'\n': true,
	'\t': true,
	'\r': true,
}

var terminalRunes = map[rune]bool{
	'\n': true,
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'^': true,
	'n': true,
	'(': true,
	')': true,
}

type IcalcLex int

func (IcalcLex) Error(s string) {
   Error("syntax error")
}

func (IcalcLex) Lex(lval *yySymType) int {
	fmt.Println("Lexing")
	var cur rune
	var i int

	cur = peekrune
	peekrune = ' '

loop:
	fmt.Println("Top of loop")
	if spaceRunes[cur] {
		cur = getrune()
		goto loop
	}
	if numberStartRunes[cur] {
		goto numb
	}
	if terminalRunes[cur] {
		return int(cur)
	}
	if cur == 0 {
		return 0
	}

	Errorf("Illegal character %c", cur)
	cur = getrune()
	goto loop

numb:
	sym = ""
	for i = 0; ; i++ {
		sym += string(cur)
		cur = getrune()
		if !numberRunes[cur] {
			break
		}
	}
	peekrune = cur
	f, err := strconv.ParseFloat(sym, 64)
	if err != nil {
		var in int64
		in, err = strconv.ParseInt(sym, 0, 32)
		if err != nil {
			Errorf("error converting %v\n", sym)
			f = 0
		} else {
			f = float64(in)
		}
	}
	lval.val = f
	return NUM
}

func readline() bool {
	fmt.Println("Reading line")
	s, err := reader.ReadString('\n')
	fmt.Println("Read line")
	fmt.Println(s)
	if err != nil {
		if err == io.EOF {
			return true
		} else {
			nerrors = 1000    // force a complete fail
			Error(err.Error())
		}
	}
	line = s
	linep = 0
	return false
}

func getrune() rune {
	fmt.Println("Getting Rune")
	if linep >= len(line) {
		if readline() {
			return 0
		}
	}

	fmt.Println("About to decode rune into string")
	cur, n := utf8.DecodeRuneInString(line[linep:len(line)])
	fmt.Printf("Decoded rune: %c", cur)
	linep += n

	if cur == '\n' {
		lineno++
	}
	return cur
}

func Errorf(s string, v ...interface{}) {
	fmt.Printf("At: %v, %v :: %v\n\t", lineno, linep, line)
	fmt.Printf(s, v...)
	fmt.Printf("\n")

	nerrors++
	if nerrors > 5 {
		fmt.Printf("too many errors\n")
		os.Exit(1)
	}
}

func Error(s string) {
	Errorf("%s", s)
}


func main() {
	reader = bufio.NewReader(os.Stdin)
	for k, v := range numberStartRunes {
		numberRunes[k] = v
	}
	for k, v := range numberMiddleOtherRunes {
		numberRunes[k] = v
	}
	for k := range terminalRunes {
		delete(spaceRunes, k)
	}
	// yyDebug = 5
	yyParse(IcalcLex(0))
}
