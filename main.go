
//line main.y:2
package main
import __yyfmt__ "fmt"
//line main.y:2
		
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


//line main.y:24
type yySymType struct{
	yys int
	num float64
	variable *variable
}

const NUM = 57346
const VAR = 57347
const FUNC = 57348
const NEG = 57349

var yyToknames = []string{
	"NUM",
	"VAR",
	"FUNC",
	" =",
	" -",
	" +",
	" *",
	" /",
	"NEG",
	" ^",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line main.y:65


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

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 16
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 57

var yyAct = []int{

	4, 17, 15, 12, 11, 13, 14, 16, 15, 18,
	19, 28, 20, 21, 22, 23, 24, 25, 26, 12,
	11, 13, 14, 2, 15, 1, 0, 27, 5, 6,
	7, 0, 8, 5, 6, 7, 0, 8, 3, 9,
	12, 11, 13, 14, 9, 15, 10, 12, 11, 13,
	14, 0, 15, 13, 14, 0, 15,
}
var yyPact = []int{

	-1000, 24, -1000, -1000, 32, -1000, 0, -14, 29, 29,
	-1000, 29, 29, 29, 29, 29, 29, 29, -11, 11,
	43, 43, -11, -11, -11, 39, -5, -1000, -1000,
}
var yyPgo = []int{

	0, 0, 25, 23,
}
var yyR1 = []int{

	0, 2, 2, 3, 3, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1,
}
var yyR2 = []int{

	0, 0, 2, 1, 2, 1, 1, 3, 4, 3,
	3, 3, 3, 2, 3, 3,
}
var yyChk = []int{

	-1000, -2, -3, 14, -1, 4, 5, 6, 8, 15,
	14, 9, 8, 10, 11, 13, 7, 15, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, 16, 16,
}
var yyDef = []int{

	1, -2, 2, 3, 0, 5, 6, 0, 0, 0,
	4, 0, 0, 0, 0, 0, 0, 0, 13, 0,
	9, 10, 11, 12, 14, 7, 0, 15, 8,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	14, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	15, 16, 10, 9, 3, 8, 3, 11, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 7, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 13,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 12,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %U %s\n", uint(char), yyTokname(c))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf("saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 4:
		//line main.y:49
		{ fmt.Printf("=> %.10g\n", yyS[yypt-1].num) }
	case 5:
		//line main.y:53
		{ yyVAL.num = yyS[yypt-0].num           }
	case 6:
		//line main.y:54
		{ yyVAL.num = yyS[yypt-0].variable.num }
	case 7:
		//line main.y:55
		{ yyVAL.num = yyS[yypt-0].num; yyS[yypt-2].variable.num = yyS[yypt-0].num }
	case 8:
		//line main.y:56
		{ yyVAL.num = yyS[yypt-3].variable.funcptr(yyS[yypt-1].num)  }
	case 9:
		//line main.y:57
		{ yyVAL.num = yyS[yypt-2].num + yyS[yypt-0].num      }
	case 10:
		//line main.y:58
		{ yyVAL.num = yyS[yypt-2].num - yyS[yypt-0].num      }
	case 11:
		//line main.y:59
		{ yyVAL.num = yyS[yypt-2].num * yyS[yypt-0].num      }
	case 12:
		//line main.y:60
		{ yyVAL.num = yyS[yypt-2].num / yyS[yypt-0].num      }
	case 13:
		//line main.y:61
		{ yyVAL.num = -yyS[yypt-0].num          }
	case 14:
		//line main.y:62
		{ yyVAL.num = math.Pow(yyS[yypt-2].num, yyS[yypt-0].num) }
	case 15:
		//line main.y:63
		{ yyVAL.num = yyS[yypt-1].num           }
	}
	goto yystack /* stack new state and value */
}
