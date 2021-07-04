package goexpr

import (
	"fmt"
	"testing"
)

func TestEngine(t *testing.T) {
	exprs := `(#3-4)<10&&4>1&&[1,2,4] Contain 4 && ADD(1,2)>1`
	eg := NewEngine()
	eg.AddFunc("ADD", 10, func(v ...interface{}) interface{} {
		return floatVal(v[0]) + floatVal(v[1])
	})
	eg.AddPrefix("#", 100, func(v interface{}) interface{} {
		return floatVal(v) * floatVal(v)
	})
	eg.AddInfix("Contain", 50, func(v1, v2 interface{}) interface{} {
		return Contain(v1, v2)
	})
	result := eg.Execute(exprs, map[string]interface{}{"ImageMode": 12, "TplIds": []float64{10025, 20}})
	fmt.Println(result)
}

func TestSpitExpr(t *testing.T) {
	exprs := `('ssss','aaa',['aaa',bb],[aaa,game notIn values],funa bsna ( sa,ssdf), atype contains (90),type notIn [1,2,4],value >= images ,-add(),[name,add()],-otEs(),[[-aaam,bb],bbb])`
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
