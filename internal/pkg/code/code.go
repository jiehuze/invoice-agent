package code

import "errors"

const (
	Success      = 200
	InvalidParam = 10000 + iota
	DBErr
	CommonError
	HTTPErr
	HTTPStatusErr
	DataNotExist
)

const (
	MsgSuccess       = "成功"
	MsgInvalidParam  = "错误的参数格式"
	MsgDBErr         = "数据库错误"
	MsgCommonError   = "失败"
	MsgHTTPErr       = "请求失败"
	MsgHTTPStatusErr = "请求非200"
	MsgDataNotExist  = "数据不存在"
)

var (
	//
	ErrHTTPStatusErr = errors.New(MsgHTTPStatusErr)
)
