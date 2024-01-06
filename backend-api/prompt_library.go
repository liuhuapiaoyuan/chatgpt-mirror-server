package backendapi

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Prompt_library(r *ghttp.Request) {

	resJson := gjson.New(`
	{
    "items": [
        {
            "id": "195cfe2a",
            "title": "构思概念",
            "description": "一个复古风格的街机游戏",
            "prompt": "为一个复古风格的街机游戏构思5个概念。"
        },
        {
            "id": "c25a3687",
            "title": "展示一个网站粘性导航的代码片段",
            "description": "使用CSS和JavaScript展示一个网站的粘性导航的代码片段",
            "prompt": "展示一个网站的粘性导航的CSS和JavaScript代码片段。"
        },
        {
            "id": "9fa376de",
            "title": "推荐一道菜",
            "description": "以迎合挑食的约会对象",
            "prompt": "我要为一个自称挑食的约会对象烹饪。你能推荐一道易于做的菜吗？"
        },
        {
            "id": "b1935e1a",
            "title": "解释这段代码：",
            "description": "\"cat config.yaml | awk NF\"",
            "prompt": "解释这个bash命令的作用：\"cat config.yaml | awk NF\""
        }
    ],
    "total": 4,
    "limit": 4,
    "offset": 0
}
`)
	r.Response.WriteJson(resJson)
}
