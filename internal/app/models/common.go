package models

type RespValue struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Err  string      `json:"err"`
	Data interface{} `json:"data"`
}

type RespInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Err  error  `json:"err"`
}

type Pager struct {
	PageNo   int `json:"page_no" form:"page_no"`
	PageSize int `json:"page_size" form:"page_size"`
}
