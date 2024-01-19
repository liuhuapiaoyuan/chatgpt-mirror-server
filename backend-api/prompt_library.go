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
            "id": "195cfe2a3",
            "title": "开题报告：",
            "description": "一个软件工程的毕业开题报告",
            "prompt": "请生成一份软件工程的开题报告，主题聚焦于‘基于云计算的敏捷软件开发流程优化’。报告应包含以下关键部分：\r\n\r\n引言：\r\n\r\n简述软件工程的当前挑战，特别是在敏捷开发领域。\r\n解释云计算在现代软件开发中的作用和潜力。\r\n研究目的和动机：\r\n\r\n明确研究的目的，即如何利用云计算优化敏捷软件开发流程。\r\n讨论此研究对软件工程领域的重要性和潜在贡献。\r\n文献综述：\r\n\r\n回顾相关领域的文献，包括敏捷开发方法、云计算技术及其在软件工程中的应用。\r\n识别并讨论现有研究的空白和限制。\r\n研究问题和假设：\r\n\r\n明确要解决的具体研究问题或要验证的假设。\r\n方法论：\r\n\r\n描述将采用的研究方法，包括数据收集和分析技术。\r\n如有可能，提出初步的研究设计和实施计划。\r\n预期结果：\r\n\r\n预测研究可能得出的结果和发现。\r\n讨论这些结果如何帮助解决敏捷软件开发中的实际问题。\r\n研究的局限性和挑战：\r\n\r\n识别可能影响研究结果的潜在局限性和挑战。\r\n时间表和里程碑：\r\n\r\n提供一个详细的研究时间表，列出重要的里程碑和预期完成日期。\r\n报告的语言应清晰、专业，确保内容全面且逻辑连贯。"
        },
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
