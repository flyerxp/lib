package vd

import (
	"fmt"
	vdbyte "github.com/bytedance/go-tagexpr/v2/validator"
)

var vdDefault *vdbyte.Validator

type VdErr struct {
	ErrType   string `json:"ErrType"`
	FailField string `json:"FailField"`
	Msg       string `json:"Msg"`
}

func (v VdErr) Error() string {
	return v.Msg
}
func (v VdErr) Json() string {
	return fmt.Sprintf(`{"ErrType":"%s","FailField":"%s","Msg":"%s"}`, v.ErrType, v.FailField, v.Msg)
}

func init() {
	vdDefault = vdbyte.New("vd").SetErrorFactory(func(failPath, msg string) error {
		return VdErr{
			ErrType:   "validating",
			FailField: failPath,
			Msg:       msg,
		}
	})
}
func VdDefault(obj any) error {
	return vdDefault.Validate(obj)
}

func VdCustom(obj any) (VdErr, error) {
	errObj := VdErr{}
	e := vdbyte.New("vd").SetErrorFactory(func(failPath, msg string) error {
		errObj = VdErr{
			ErrType:   "validating",
			FailField: failPath,
			Msg:       msg,
		}

		return errObj
	}).Validate(obj)
	return errObj, e
}
