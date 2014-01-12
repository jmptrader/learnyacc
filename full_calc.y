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

type mathFunc func(float64) float64

type variable struct {
	kind int       // either VAR or FUNC
	num float64
	funcptr mathFunc
}

%}

%union{
	num float64
	variable *variable
}

%token <num> NUM
%token <variable> VAR FUNC

%type <num> exp

%right '='
%left '-' '+'
%left '*' '/'
%left NEG		// negation--unary minus
%right '^'			// exponentiation

%%

input:
	  /* empty */
	| input line


line:
	  '\n'
	| exp '\n'		{ fmt.Printf("=> %.10g\n", $1) }


exp:
	  NUM				{ $$ = $1           }
	| VAR				{ $$ = $1.num }
	| VAR '=' exp		{ $$ = $3; $1.num = $3 }
	| FUNC '(' exp ')'	{ $$ = $1.funcptr($3)  }
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

var numberStartRunes map[rune]bool
var numberRunes map[rune]bool
var identifierStartRunes map[rune]bool
var identifierRunes map[rune]bool
var spaceRunes map[rune]bool
var terminalRunes map[rune]bool

func initRuneMaps() {
	var digits = mapToTrue("0123456789")
	var alpha = mapToTrue("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var identifierOther = mapToTrue("_")
	var otherNumberStartRunes = mapToTrue(".")
	var numberMiddleOtherRunes = mapToTrue("eExX-+")

	// Any terminalRunes will be deleted from this
	var spaceRuneCandidates = mapToTrue(" \n\r\t")

	// terminalRunes
	terminalRunes = mapToTrue("\n+-*/^n()=")

	// numberStartRunes
	numberStartRunes = combinedMaps(digits, otherNumberStartRunes)

	// numberRunes
	numberRunes = combinedMaps(numberStartRunes, numberMiddleOtherRunes)

	// spaceRunes
	for k := range terminalRunes {
		delete(spaceRuneCandidates, k)
	}
	spaceRunes = spaceRuneCandidates

	// identifierStartRunes
	identifierStartRunes = combinedMaps(alpha, identifierOther)

	// identifierRunes
	identifierRunes = combinedMaps(identifierStartRunes, digits)
}

func initSymbolTable() {
	var initialFunctions = map[string]func (float64) float64 {
		"abs": math.Abs,
		"exp": math.Exp,
		"log": math.Log,
		"log2": math.Log2,
		"cos": math.Cos,
		"sin": math.Sin,
	}

	for k, v := range initialFunctions {
		symbolTable[k] = &variable{kind: FUNC, funcptr: mathFunc(v)}
	}
}

func mapToTrue(chars string) map[rune]bool {
	var m = make(map[rune]bool)
	for _, r := range chars {
		m[r] = true
	}
	return m
}

func mergeMap(dest map[rune]bool, sources... map[rune]bool) map[rune]bool {
	for _, s := range sources {
		for k, v := range s {
			dest[k] = v
		}
	}
	return dest
}

func combinedMaps(maps... map[rune]bool) map[rune]bool {
	return mergeMap(make(map[rune]bool), maps...)
}

var symbolTable = make(map[string]*variable)

func getVariable(sym string) *variable {
	v, ok := symbolTable[sym]
	if ok {
		return v
	}
	v = &variable{kind: VAR, num: 0}
	symbolTable[sym] = v
	return v
}

type MulticalcLex int

func (MulticalcLex) Error(s string) {
   Error("syntax error")
}

func (MulticalcLex) Lex(lval *yySymType) int {
	var cur rune
	var i int

	cur = peekrune
	peekrune = ' '

loop:
	if spaceRunes[cur] {
		cur = getrune()
		goto loop
	}
	if numberStartRunes[cur] {
		goto numb
	}
	if identifierStartRunes[cur] {
		goto identifier
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

identifier:
	sym = ""
	for i = 0; ; i++ {
		sym += string(cur)
		cur = getrune()
		if !identifierRunes[cur] {
			break
		}
	}
	peekrune = cur
	lval.variable = getVariable(sym)
	return lval.variable.kind

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
	lval.num = f
	return NUM
}

func readline() bool {
	s, err := reader.ReadString('\n')
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
	if linep >= len(line) {
		if readline() {
			return 0
		}
	}

	cur, n := utf8.DecodeRuneInString(line[linep:len(line)])
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
	initRuneMaps()
	initSymbolTable()

	// yyDebug = 5
	yyParse(MulticalcLex(0))
}
