package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

//ParseRequestBody 解析gin的POST Body
func parseRequestBody(c *gin.Context, request interface{}) (err error) {
	var data []byte
	postData, has := c.Get("postdata")
	if has {
		data = postData.([]byte)
	} else {
		data, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return
		}
		c.Set("postdata", data)
	}
	fmt.Printf("request:%s", data)
	err = json.Unmarshal(data, &request)
	if err != nil {
		return err
	}
	return err
}

func responseJSON(c *gin.Context, code Code, args ...interface{}) {
	body := make(map[string]interface{})
	body["code"] = code
	for i := 0; i < len(args); i += 2 {

		switch args[i].(type) {
		case string:
			body[args[i].(string)] = args[i+1]
			break
		}
	}
	c.JSON(http.StatusOK, body)
}

func isEmpty(args ...interface{}) bool {
	for _, arg := range args {
		switch arg.(type) {
		case int:
			if arg.(int) == 0 {
				return true
			}
		case int32:
			if arg.(int32) == 0 {
				return true
			}
		case int64:
			if arg.(int64) == 0 {
				return true
			}
		case string:
			if arg.(string) == "" {
				return true
			}
		case float64:
			if arg.(float64) == 0 {
				return true
			}
		case float32:
			if arg.(float64) == 0 {
				return true
			}
		}
	}
	return false
}
