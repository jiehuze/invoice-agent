package util

import "encoding/json"

func GetJson(v interface{}) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}
