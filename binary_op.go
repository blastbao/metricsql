package metricsql

import (
	"fmt"
	"math"
	"strings"

	"github.com/blastbao/metricsql/binaryop"
)

// 二元操作符
var binaryOps = map[string]bool{

	// 数学运算
	"+": true,
	"-": true,
	"*": true,
	"/": true,
	"%": true,
	"^": true,

	// 比较
	"==": true,
	"!=": true,
	">":  true,
	"<":  true,
	">=": true,
	"<=": true,

	// 逻辑
	"and":    true,
	"or":     true,
	"unless": true,

	// 条件
	"if":      true,
	"ifnot":   true,
	"default": true,
}

// 二元操作符优先级
var binaryOpPriorities = map[string]int{

	"default": -1,

	"if":    0,
	"ifnot": 0,

	// See https://prometheus.io/docs/prometheus/latest/querying/operators/#binary-operator-precedence
	"or":     1,
	"and":    2,
	"unless": 2,

	"==": 3,
	"!=": 3,
	"<":  3,
	">":  3,
	"<=": 3,
	">=": 3,

	"+": 4,
	"-": 4,

	"*": 5,
	"/": 5,
	"%": 5,

	"^": 6,
}

func isBinaryOp(op string) bool {
	op = strings.ToLower(op)
	return binaryOps[op]
}

func binaryOpPriority(op string) int {
	op = strings.ToLower(op)
	return binaryOpPriorities[op]
}

// 检查 s 是否前缀包含某操作符（最长匹配原则），若包含，返回对应操作符的字符数
func scanBinaryOpPrefix(s string) int {
	n := 0
	// 遍历所有二元操作符
	for op := range binaryOps {

		// 因为 s 包含 op，所以若 len(op) > len(s) ，则忽略当前 op
		if len(op) > len(s) {
			continue
		}

		// 至此，len(s) >= len(op)，所以取出 s[:len(op)] 和 op 进行比对
		ss := strings.ToLower(s[:len(op)])

		// 如果相匹配，则 op 为 ss 的前缀，更新 n = len(op) > n ? len(op) : n
		if ss == op && len(op) > n {
			n = len(op)
		}
	}

	// 至此，若 n != 0，则找到了满足最长匹配原则的操作符
	return n
}

func isRightAssociativeBinaryOp(op string) bool {
	// See https://prometheus.io/docs/prometheus/latest/querying/operators/#binary-operator-precedence
	return op == "^"
}

func isBinaryOpGroupModifier(s string) bool {
	s = strings.ToLower(s)
	switch s {
	// See https://prometheus.io/docs/prometheus/latest/querying/operators/#vector-matching
	case "on", "ignoring":
		return true
	default:
		return false
	}
}

func isBinaryOpJoinModifier(s string) bool {
	s = strings.ToLower(s)
	switch s {
	case "group_left", "group_right":
		return true
	default:
		return false
	}
}

func isBinaryOpBoolModifier(s string) bool {
	s = strings.ToLower(s)
	return s == "bool"
}

// IsBinaryOpCmp returns true if op is comparison operator such as '==', '!=', etc.
func IsBinaryOpCmp(op string) bool {
	switch op {
	case "==", "!=", ">", "<", ">=", "<=":
		return true
	default:
		return false
	}
}

func isBinaryOpLogicalSet(op string) bool {
	op = strings.ToLower(op)
	switch op {
	case "and", "or", "unless":
		return true
	default:
		return false
	}
}

func binaryOpEval(op string, left, right float64, isBool bool) float64 {


	var result float64

	// compare ?
	if IsBinaryOpCmp(op) {

		//
		evalCmp := func(cf func(left, right float64) bool) float64 {
			if isBool {
				if cf(left, right) {
					return 1
				}
				return 0
			}
			if cf(left, right) {
				return left
			}
			return nan
		}

		switch op {
		case "==":
			result = evalCmp(binaryop.Eq)
		case "!=":
			result = evalCmp(binaryop.Neq)
		case ">":
			result = evalCmp(binaryop.Gt)
		case "<":
			result = evalCmp(binaryop.Lt)
		case ">=":
			result = evalCmp(binaryop.Gte)
		case "<=":
			result = evalCmp(binaryop.Lte)
		default:
			panic(fmt.Errorf("BUG: unexpected comparison binaryOp: %q", op))
		}
	} else {
		switch op {
		case "+":
			result = binaryop.Plus(left, right)
		case "-":
			result = binaryop.Minus(left, right)
		case "*":
			result = binaryop.Mul(left, right)
		case "/":
			result = binaryop.Div(left, right)
		case "%":
			result = binaryop.Mod(left, right)
		case "^":
			result = binaryop.Pow(left, right)
		case "and":
			// Nothing to do
		case "or":
			// Nothing to do
		case "unless":
			result = nan
		case "default":
			result = binaryop.Default(left, right)
		case "if":
			result = binaryop.If(left, right)
		case "ifnot":
			result = binaryop.Ifnot(left, right)
		default:
			panic(fmt.Errorf("BUG: unexpected non-comparison binaryOp: %q", op))
		}
	}

	return result
}

var nan = math.NaN()
