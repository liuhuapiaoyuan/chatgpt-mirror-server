package api

import (
	"chatgpt-mirror-server/config"
	"net/http"
	"strings"

	"github.com/cool-team-official/cool-admin-go/cool"
	basemodel "github.com/cool-team-official/cool-admin-go/modules/base/model"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Share(r *ghttp.Request) {
	ctx := r.GetCtx()
	UpStream := config.CHATPROXY(ctx)
	// 分享按钮
	buttonContent := "Get started with <span class=\"font-bold\">ChatGPT</span>"

	shareButtonRecord, err := cool.DBM(basemodel.NewBaseSysParam()).WhereLike("keyName", "chatgpt.share_button_text").One()
	if err == nil && shareButtonRecord != nil {
		buttonContent = shareButtonRecord["data"].String()
	}
	shareId := r.GetRouter("shareId").String()
	res, err := g.Client().SetHeaderMap(map[string]string{
		"User-Agent": r.Header.Get("User-Agent"),
	}).Get(ctx, UpStream+"/share/"+shareId)
	if err != nil {
		r.Response.WriteStatus(res.StatusCode)
		return
	}
	resStr := res.ReadAllString()
	// 如果包含502
	if strings.Contains(resStr, "502") {
		r.Response.WriteStatus(http.StatusBadGateway)
		return
	}
	g.Log().Debug(ctx, "share", resStr)

	// 截取 <body class="antialiased"><div id="__next"> 后面的所有文本
	infoContainerStr := "<div class=\"w-full pt-2 md:pt-0 dark:border-white/20 md:border-transparent md:dark:border-transparent md:w-[calc(100%-.5rem)]\">"
	insertStr := "<div class=\"relative flex w-full flex-1 items-center justify-center gap-2 pt-3 empty:hidden\"><a class=\"flex items-center gap-2 rounded-xl bg-green-600 px-5 py-3.5 font-medium text-white transition hover:opacity-70\" href=\"/\"><div>" + buttonContent + "</div></a></div>"
	bodyStr := resStr[strings.Index(resStr, "<body class=\"antialiased\"><div id=\"__next\">")+len("<body class=\"antialiased\"><div id=\"__next\">"):]

	// 替换bodyStr的infoContainerStr为 infoContainerStr+insertStr
	insertPos := strings.Index(bodyStr, infoContainerStr) + len(infoContainerStr)
	bodyStr = bodyStr[:insertPos] + insertStr + bodyStr[insertPos:]

	// 暂时冻结share
	r.Response.WriteTpl("/share-20231119.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(ctx),
		"content":     bodyStr,
	})
}
