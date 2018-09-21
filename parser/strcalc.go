package parser

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

/*
###########################
# Section: Structs
###########################
*/

type Stack struct {
	Items  []float64
	Length int
}

// pushes an item onto the stack
func (s *Stack) Push(item float64) {
	s.Items = append(s.Items, item)
	s.Length++
}

// pops an item from the stack
func (s *Stack) Pop() (item float64) {
	s.Length--
	item = s.Items[s.Length]
	s.Items = s.Items[:s.Length]
	return
}

func (s *Stack) PerformOperation(op rune) {
	var res float64

	var2 := s.Pop()
	var1 := s.Pop()

	// apply operation
	switch op {
	case '+':
		res = var1 + var2
	case '-':
		res = var1 - var2
	case '*':
		res = var1 * var2
	case '/':
		res = var1 / var2
	case '^':
		res = math.Pow(var1, var2)
	}

	// push the operation result to the value stack
	s.Push(res)
}

/*
########################
# Section: Code
########################
*/

// Regular expressions
var alphaNumericRegex *regexp.Regexp
var functionRegex *regexp.Regexp

var prefixValueRegex *regexp.Regexp
var suffixValueRegex *regexp.Regexp

// Plugin struct for communicating with the app.so file
var plug *Plugin

func Init(p *Plugin) {
	plug = p

	alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+`)
	functionRegex = regexp.MustCompile(`^[a-zA-Z0-9]+\((\$[a-zA-Z0-9], ?)*(\$[a-zA-Z0-9])?\)`)

	prefixValueRegex = regexp.MustCompile(`^((\(.*\))|([0-9]+)|([0-9]+\.[0-9]+))`)
	suffixValueRegex = regexp.MustCompile(`((\(.*\))|([0-9]+)|([0-9]+\.[0-9]+))$`)
}

// Calculates a integer value of a string, e.g
// "(4/3)*@GetPi()*($r^3)"
func CalculateValue(op string) (result string) {
	// temporarely remove % suffix if there is one
	isPercentage := false
	if strings.HasSuffix(op, "%") {
		op = op[:len(op)-1]
		isPercentage = true
	}

	// in case of variables or functions evaluate them.
	for i, c := range op {
		prev := op[:i+1]
		if c == '@' {
			op = prev + evalNextFunction(op[i+1:])
		} else if c == '$' {
			op = prev + evalNextVariable(op[i+1:])
		}
	}

	// build the command stack
	// the last item should be a value
	// containing the result
	flresult := buildStack(op).Pop()

	// convert the result to a string
	result = strconv.Itoa(int(math.Round(flresult)))

	// reattach percentage sign
	if isPercentage {
		result = result + "%"
	}
	return
}

// builds a stack from a string containing operations
// (not variables or functions!)
func buildStack(op string) (result *Stack) {
	// initialize result
	result = &Stack{[]float64{}, 0}

	// "bracketize":
	// a*b*c -> (a*b)*c
	i := -1
	for i != 0 {
		i = getOperandIndex(op)

		leftVal := suffixValueRegex.FindString(op[:i])
		rightVal := prefixValueRegex.FindString(op[i+1:])

		if len(leftVal) > 0 && len(rightVal) > 0 {
			// append brackets
			op = op[:i-len(leftVal)] + "(" +
				op[i-len(leftVal):i+len(rightVal)+1] + ")" +
				op[i+len(rightVal)+1:]
		}
	}

	// get rid of outer brackets
	// while the first bracket matches the last index
	for getMatchingBracketIndex(op, 0) == len(op)-1 && len(op) > 2 {
		op = op[1 : len(op)-1]
	}

	// evaluate operation
	i = getOperandIndex(op)

	leftVal := suffixValueRegex.FindString(op[:i])
	rightVal := prefixValueRegex.FindString(op[i+1:])

	result.Push(evalValue(leftVal))
	result.Push(evalValue(rightVal))

	result.PerformOperation([]rune(op)[i])

	return
}

// gets the next operand which is not in brackets, sorted by operand tier
func getOperandIndex(op string) (index int) {
	// get tier 3 operands
	for i, c := range op {
		if c == '^' {
			if !isOpInBrackets(op[i:]) {
				return i
			}
		}
	}

	// get tier 2 operands
	for i, c := range op {
		if c == '*' || c == '/' {
			if !isOpInBrackets(op[i:]) {
				return i
			}
		}
	}

	// get tier 1 operands
	for i, c := range op {
		if c == '+' || c == '-' {
			if !isOpInBrackets(op[i:]) {
				return i
			}
		}
	}

	// nothing was found
	return 0
}

// evaluates a value (or operations) and puts the result on given stack
func evalValue(value string) (result float64) {
	// if value is enclosed in brackets
	if strings.HasPrefix(value, "(") {
		// build the stack of the expression without the surrounding brackets
		valStack := buildStack(value[1 : len(value)-1])

		return valStack.Pop()
	} else {
		// cast the value to float64 and put it onto the stack
		floatval, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal(err)
			return
		}
		return floatval
	}
}

// gets if the first char in an equation is in brackets
func isOpInBrackets(abOp string) (isInBrackets bool) {
	// opening brackets - closing brackets after abOp[ind]
	openBracketCount := 0
	for _, c := range abOp {
		if c == '(' {
			openBracketCount++
		} else if c == ')' {
			openBracketCount--
		}
	}

	// if more were closed than opened,
	// there must be an 'active' opening bracket
	// before ind
	if openBracketCount < 0 {
		return true
	} else {
		return false
	}
}

// gets the matching bracket of bracket br
func getMatchingBracketIndex(op string, brindex int) (index int) {
	openBracketCount := 0
	for i, c := range op[brindex:] {
		if c == '(' {
			openBracketCount++
		} else if c == ')' {
			openBracketCount--
		}

		// if the bracket count is 0, it means the bracket requested
		// was closed (or never opened)
		if openBracketCount == 0 {
			return i
		}
	}

	// nothing was found
	return 0
}

// evaluates the variable at pos 0 in given string and returns
// the rest of the string
func evalNextVariable(str string) (result string) {
	varname := alphaNumericRegex.FindString(str)
	return plug.GetVariable(varname) + str[len(varname):]
}

// evaluates the function at pos 0 in given string and returns
// the rest of the string
func evalNextFunction(str string) (result string) {
	funcname := functionRegex.FindString(str)
	return plug.CallFunction(funcname) + str[len(funcname):]
}
