package backendapi

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Prompt_library(r *ghttp.Request) {

	resJson := gjson.New(`{
        "items": [
            {
                "id": "e3c32040",
                "title": "为我创建个人网页",
                "description": "先问我 3 个问题",
                "oneliner": "为我创建个人网页",
                "prompt": "为我创建个人网页，需要全部包含在一个文件中。先就你想知道的内容问我 3 个问题。",
                "category": "code"
            },
            {
                "id": "38034ee7",
                "title": "解释超导体",
                "description": "就好比我是一个五岁小孩",
                "oneliner": "解释超导体",
                "prompt": "把我当做五岁小朋友一样，向我解释超导体。",
                "category": "teach-or-explain"
            },
            {
                "id": "f2251c8d",
                "title": "给我提供一些主意",
                "description": "关于如何处理孩子们的艺术作品",
                "oneliner": "如何处理孩子们的艺术作品",
                "prompt": "我可以用孩子们的艺术作品做哪五种富有创意的事？我不想把它们扔掉，但也太乱了。",
                "category": "idea"
            },
            {
                "id": "a477416b",
                "title": "告诉我一个趣事",
                "description": "关于罗马帝国",
                "oneliner": "关于罗马帝国的趣事",
                "prompt": "随便告诉我一个关于罗马帝国的趣事",
                "category": "misc"
            }
        ],
        "total": 4,
        "limit": 4,
        "offset": 0
    }`)
	r.Response.WriteJson(resJson)
}
