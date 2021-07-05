package goexpr

import (
	"bytes"
	"fmt"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Engine struct {
	priority    map[string]int32
	prefixSet   map[string]PrefixOp
	infixSet    map[string]InfixOp
	functionSet map[string]FunctionOp
	funcArgs    map[string]int
	operaSet    []string
}

func NewEngine() *Engine {
	prefixSet := map[string]PrefixOp{}
	priority := map[string]int32{}
	infixSet := map[string]InfixOp{}
	functionSet := map[string]FunctionOp{}
	operaSet := []string{"(", ")", "[", "]", ","}
	for k, v := range PrefixOpSet {
		prefixSet[k] = v
		operaSet = append(operaSet, k)
	}
	for k, v := range InfixOpSet {
		infixSet[k] = v
		operaSet = append(operaSet, k)
	}
	for k, v := range OpPriority {
		priority[k] = v
	}
	return &Engine{
		prefixSet:   prefixSet,
		priority:    priority,
		infixSet:    infixSet,
		functionSet: functionSet,
		operaSet:    operaSet,
		funcArgs:    map[string]int{},
	}
}

func (en *Engine) AddFunc(fname string, priority int32, op FunctionOp) {
	en.functionSet[fname] = op
	en.priority[fname] = priority
	en.operaSet = append(en.operaSet, fname)
}

func (en *Engine) AddPrefix(fname string, priority int32, op PrefixOp) {
	en.prefixSet[fname] = op
	en.priority[fname] = priority
	en.operaSet = append(en.operaSet, fname)
}

func (en *Engine) AddInfix(fname string, priority int32, op InfixOp) {
	en.infixSet[fname] = op
	en.priority[fname] = priority
	en.operaSet = append(en.operaSet, fname)
}

func (en *Engine) Execute(expression string, args map[string]interface{}) interface{} {
	exprs := en.expression(expression)
	fmt.Println(exprs)
	numbs := lls.New()
	operas := lls.New()
	for _, exprStr := range exprs {
		if numb, ok := GetNumber(exprStr); ok {
			numbs.Push(numb)
			continue
		}
		if exprStr != "'" && hasPreSufix(exprStr, "'", "'") {
			numbs.Push(exprStr[1 : len(exprStr)-1])
			continue
		}
		if len(exprStr) > 1 && strings.HasPrefix(exprStr, "(") {
			exprList := SpitExpr(exprStr)
			top, _ := operas.Peek()
			if top != nil && en.functionSet[top.(string)] != nil {
				en.funcArgs[top.(string)] = len(exprList)
			}
			for _, tempExpr := range exprList {
				numb := en.Execute(tempExpr, args)
				numbs.Push(numb)
			}
			continue
		}
		if strings.HasPrefix(exprStr, "[") {
			var array []interface{}
			exprList := SpitExpr(exprStr)
			for _, tempExpr := range exprList {
				numb := en.Execute(tempExpr, args)
				array = append(array, numb)
			}
			numbs.Push(array)
			continue
		}
		if exprStr == ")" {
			//计算括号内部的,直到计算到(
			en.CalculateBract(operas, numbs)
			continue
		}
		if Has(en.operaSet, exprStr) {
			en.PushCurOpera(exprStr, operas, numbs)
			continue
		}
		numbs.Push(GetArg(exprStr, args))
	}
	en.CalculateStack(operas, numbs)
	result, _ := numbs.Pop()
	return result
}

func hasPreSufix(exprs string, s string, e string) bool {
	return strings.HasPrefix(exprs, s) && strings.HasSuffix(exprs, e)
}

func GetArg(path string, args map[string]interface{}) interface{} {
	idx := strings.Index(path, ".")
	if idx < 0 {
		return args[path]
	}
	if args[path[:idx]] == nil {
		return nil
	}
	tmpArgs, ok := args[path[:idx]].(map[string]interface{})
	if !ok {
		return nil
	}
	return GetArg(path[idx+1:], tmpArgs)
}

// 23+46*56-5*Add(-4-6) IN [1,2,3+4]
func (eng *Engine) expression(exprs string) []string {
	var idx = 0
	var exprLen = len(exprs)
	var exprList []string
	sort.Slice(eng.operaSet, func(i, j int) bool {
		return len(eng.operaSet[i]) > len(eng.operaSet[j])
	})
	for {
		if idx >= exprLen {
			break
		}
		item := rune(exprs[idx])
		if unicode.IsSpace(item) {
			idx++
			continue
		}
		preFix := ""
		if len(exprList) > 0 {
			preFix = exprList[len(exprList)-1]
		}
		if exprs[idx] == '-' && eng.IsNgvFlag(preFix) {
			idx += 1
			exprList = append(exprList, Ngv)
			continue
		}
		reg, _ := regexp.Compile(`^\s*[0-9]+\.*[0-9]*`)
		if numb := reg.FindString(exprs[idx:]); len(numb) > 0 {
			idx += len(numb)
			numb = strings.ReplaceAll(numb, " ", "")
			exprList = append(exprList, numb)
			continue
		}
		if eng.functionSet[preFix] != nil {
			argExpr := match(exprs[idx:], "(", ")")
			idx += len(argExpr)
			exprList = append(exprList, argExpr)
			continue
		}
		if string(exprs[idx]) == "[" {
			array := match(exprs[idx:], "[", "]")
			idx += len(array)
			exprList = append(exprList, array)
			continue
		}
		if strings.HasPrefix(exprs[idx:], "'") {
			end := strings.Index(exprs[idx+1:], "'")
			str := exprs[idx : idx+end+2]
			idx += len(str)
			exprList = append(exprList, str)
			continue
		}
		reg, _ = regexp.Compile(`^[0-9]+\.*[0-9]*`)
		if numb := reg.FindString(exprs[idx:]); len(numb) > 0 {
			idx += len(numb)
			exprList = append(exprList, numb)
			continue
		}
		var opera = ""
		reg, _ = regexp.Compile(`^[^A-Za-z0-9_]`)
		reg1, _ := regexp.Compile(`[A-Za-z][A-Za-z0-9]*`)
		for _, op := range eng.operaSet {
			if !strings.HasPrefix(exprs[idx:], op) {
				continue
			}
			if reg1.MatchString(op) && !reg.MatchString(exprs[len(op):]) {
				continue
			}
			opera = op
			break
		}
		if len(opera) > 0 {
			if opera == "-" && !eng.IsNgvFlag(preFix) {
				opera = Sub
			}
			idx += len(opera)
			exprList = append(exprList, opera)
			continue
		}
		reg, _ = regexp.Compile(`^[A-Za-z][A-za-z0-9]*[,\s\)\]]`)
		if variable := reg.FindString(exprs[idx:]); len(variable) > 0 {
			reg1, _ := regexp.Compile(`[,\s\)\]]*$`)
			variable = reg1.ReplaceAllString(variable, "")
			idx += len(variable)
			exprList = append(exprList, variable)
			continue
		}
		reg, _ = regexp.Compile(`^[A-Za-z][A-za-z0-9]*[,\s\)\]]$`)
		if variable := reg.FindString(exprs[idx:]); len(variable) > 0 {
			reg1, _ := regexp.Compile(`[,\s\)\]]*$`)
			variable = reg1.ReplaceAllString(variable, "")
			idx += len(variable)
			exprList = append(exprList, variable)
			continue
		}
		var minOpIdx = -1
		for _, op := range eng.operaSet {
			opIdx := strings.Index(exprs[idx:], op)
			if opIdx < 0 {
				continue
			}
			if reg1.MatchString(op) {
				reg2, _ := regexp.Compile(fmt.Sprintf(`[^A-Za-z]?%v[^A-Za-z]`, op))
				begin := idx + opIdx
				if idx > 0 {
					begin = begin - 1
				}
				if !reg2.MatchString(exprs[begin:]) {
					continue
				}
			}
			if (minOpIdx == -1 || opIdx < minOpIdx) || (opIdx == minOpIdx && opIdx >= 0 && len(opera) < len(op)) {
				minOpIdx = opIdx
				opera = op
			}
		}
		if len(opera) > 0 {
			variable := exprs[idx : idx+minOpIdx]
			exprList = append(exprList, variable)
			idx += len(variable)
			continue
		}
		exprList = append(exprList, strings.TrimSpace(exprs[idx:]))
		break
	}
	return exprList
}

func (eng *Engine) IsNgvFlag(exprStr string) bool {
	return exprStr == "(" || exprStr == "" || exprStr == "[" || exprStr == "," || eng.infixSet[exprStr] != nil || eng.prefixSet[exprStr] != nil
}

// ('ssss','aaa',['aaa',bb],[aaa,game notIn values],funa bsna ( sa,ssdf), atype contains (90),type notIn [1,2,4],value >= images ,-add(),[name,add()],-otEs(),[[-aaam,bb],bbb])
func SpitExpr(exprs string) []string {
	exprs = exprs[1 : len(exprs)-1]
	var exprList []string
	idx := 0
	for {
		if idx >= len(exprs) {
			break
		}
		buf := &bytes.Buffer{}
		jdx := idx
		for {
			if jdx >= len(exprs) {
				break
			}
			if exprs[jdx] == ',' {
				jdx++
				break
			}
			if exprs[jdx] == '[' {
				exprStr := match(exprs[jdx:], "[", "]")
				jdx += len(exprStr)
				buf.WriteString(exprStr)
				continue
			}
			if exprs[jdx] == '(' {
				exprStr := match(exprs[jdx:], "(", ")")
				jdx += len(exprStr)
				buf.WriteString(exprStr)
				continue
			}
			buf.WriteByte(exprs[jdx])
			jdx++
		}
		exprStr := strings.TrimSpace(buf.String())
		exprList = append(exprList, exprStr)
		idx = jdx
	}
	return exprList
}

func GetNumber(str string) (float64, bool) {
	number, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return number, true
	}
	return 0, false
}

func (eng *Engine) CalculateStack(opStack, nbStack *lls.Stack) {
	for {
		top, _ := opStack.Peek()
		if top == nil {
			break
		}
		eng.CalculateTop(opStack, nbStack)
	}
}

func (eng *Engine) CalculateBract(opStack, nbStack *lls.Stack) {
	for {
		top, _ := opStack.Peek()
		if top == BraktLeft {
			opStack.Pop()
			break
		}
		if top == nil {
			panic("you expr miss left (")
		}
		eng.CalculateTop(opStack, nbStack)
	}
}

func (eng *Engine) PushCurOpera(curOp string, opStack, nbStack *lls.Stack) {
	if opStack.Empty() {
		opStack.Push(curOp)
		return
	}
	for {
		top, _ := opStack.Peek()
		if top == nil {
			opStack.Push(curOp)
			break
		}
		topOp := top.(string)
		if topOp == BraktLeft || topOp == ArrayLeft {
			opStack.Push(curOp)
			break
		}
		if eng.prefixSet[curOp] != nil {
			opStack.Push(curOp)
			break
		}
		if eng.functionSet[curOp] != nil {
			opStack.Push(curOp)
			break
		}
		topPty := eng.priority[topOp]
		curPty := eng.priority[curOp]
		if topPty > curPty {
			opStack.Push(curOp)
			break
		}
		eng.CalculateTop(opStack, nbStack)
	}
}

func (eng *Engine) CalculateTop(opStack, nbStack *lls.Stack) {
	top, _ := opStack.Peek()
	if top == nil {
		return
	}
	if fun, ok := eng.infixSet[top.(string)]; ok {
		numb1, _ := nbStack.Pop()
		numb2, _ := nbStack.Pop()
		result := fun(numb2, numb1)
		nbStack.Push(result)
		opStack.Pop()
		return
	}
	if fun, ok := eng.prefixSet[top.(string)]; ok {
		numb1, _ := nbStack.Pop()
		result := fun(numb1)
		nbStack.Push(result)
		opStack.Pop()
		return
	}
	if fun, ok := eng.functionSet[top.(string)]; ok {
		argCount := eng.funcArgs[top.(string)]
		var params = make([]interface{}, argCount)
		for i := 0; i < argCount; i++ {
			numb, _ := nbStack.Pop()
			params[argCount-i-1] = numb
		}
		result := fun(params...)
		nbStack.Push(result)
		opStack.Pop()
		return
	}
	panic(fmt.Sprintf("No find function '%v'", top))
}

func match(va string, left, right string) string {
	lCount := 0
	rCount := 0
	for idx, v := range va {
		if string(v) == left {
			lCount++
		}
		if string(v) == right {
			rCount++
		}
		if lCount == rCount {
			return va[:idx+1]
		}
	}
	if lCount > 0 {
		panic("expr is miss right " + right)
	}
	return ""
}

func Has(array []string, va string) bool {
	for _, v := range array {
		if v == va {
			return true
		}
	}
	return false
}
