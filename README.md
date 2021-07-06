# goexpr
golang 表达式引擎，规则引擎，支持自定义函数，自定义中缀操作符，自定义前缀操作符，支持传参以及参数层级嵌套
```
import (
	"fmt"
	"testing"
)

func TestEngine(t *testing.T) {
	exprs := `4+3>5&&5<4+5&&3NotIN[1,2,4]&&ADD(1,2)<4&&-(#-3-4)<=30&&4>1&&[1,2,4] Contain 4 && ADD(1,2)!=1 && user.name=='kiteee' && user_count>20`
	//exprs = `-------1`
	//exprs=`user.name=='kiteee' && user_count>20`
	//exprs=`user_count>20 && user_count>20`
	//exprs=`#--3*-4-#2`
	//exprs=`-4-#2`
	//exprs = `3NotIN([1,2,3])&&ADD(1,2)<4`
	//exprs = `[1,2,4] Contain 4 IN [true] NotIN [false]`
	eg := NewEngine()
	eg.AddFunc("ADD", func(v ...interface{}) interface{} {
		return floatVal(v[0]) + floatVal(v[1])
	})
	eg.AddPrefix("#", func(v interface{}) interface{} {
		return floatVal(v) * floatVal(v)
	})
	eg.AddInfix("Contain", 30, func(v1, v2 interface{}) interface{} {
		return Contain(v1, v2)
	})
	var params = map[string]interface{}{
		"user": map[string]interface{}{
			"name": "kiteee",
			"age":  50,
		},
		"user_count": 30,
	}
	//eg.SetPriority("NotIN", 30)
	result := eg.Execute(exprs, params)
	fmt.Println(result)
}

func TestSpitExpr(t *testing.T) {
	exprs := `('ssss','aaa',['aaa',bb],[aaa,game notIn values],funa bsna ( sa,ssdf), atype contains (90),type notIn [1,2,4],value >= images ,-add(),[name,add()],-otEs(),[[-aaam,bb],bbb])`
	exprs = `()`
	result := SpitExpr(exprs)
	for _, v := range result {
		fmt.Println(v)
	}
}

func TestGetArgs(t *testing.T) {
	var mp = map[string]interface{}{
		"user": map[string]interface{}{
			"name": "kiteee",
			"age":  50,
		},
	}
	va := GetArg("user.name", mp)
	fmt.Println(va)
}
```
