package utils

import "encoding/json"

//DingDingMessage ...
func DingDingMessage(message string) {
	hookURL := "https://oapi.dingtalk.com/robot/send?access_token=733577036f50bbea9f26f8d3da44a1be467a41f0c2dc99ea1da427224fe5942a"
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": "BFSS:" + message,
		},
	}
	body, _ := json.Marshal(&msg)
	HTTPPost(hookURL, body, nil)
}
