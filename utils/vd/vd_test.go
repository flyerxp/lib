package vd

import (
	"fmt"
	"testing"
)

func TestVd(T *testing.T) {
	type InfoRequest struct {
		Name         string   `vd:"($!='Alice'||(Age)$==18) && regexp('\\w')"`
		Age          int      `vd:"$>0"`
		Email        string   `vd:"email($)"`
		Phone1       string   `vd:"phone($)"`
		OtherPhones  []string `vd:"range($, phone(#v,'CN'))"`
		*InfoRequest `vd:"?"`
		Info1        *InfoRequest `vd:"?"`
		Info2        *InfoRequest `vd:"-"`
	}
	info := &InfoRequest{
		Name:        "Alice",
		Age:         18,
		Email:       "henrylee2cn@gmail.com",
		Phone1:      "+8618812345678",
		OtherPhones: []string{"18812345679", "18812345680"},
	}
	fmt.Println(VdCustom(info))

	type A struct {
		A    int `vd:"$<0||$>=100"`
		Info interface{}
	}
	info.Email = "xxx"
	a := &A{A: 107, Info: info}
	fmt.Println(VdCustom(a))
	type B struct {
		B string `vd:"len($)>1 && regexp('^\\w*$')"`
	}
	b := &B{"abc"}
	fmt.Println(VdDefault(b) == nil)
	type C struct {
		C bool `vd:"@:(S.A)$>0 && !$; msg:'C must be false when S.A>0'"`
		S *A
	}
	c := &C{C: true, S: a}
	obj, err := VdCustom(c)
	fmt.Println(obj.Json(), err)
}
