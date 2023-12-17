package tasks

import (
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/model"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
)

func init() {
	ctx := gctx.GetInitCtx()
	// 任务注册
	corn, err := gcron.AddSingleton(ctx, config.CRONINTERVAL(ctx), RefreshSession, "RefreshSession")
	if err != nil {
		panic(err)
	}
	g.Log().Info(ctx, "RefreshSession", "corn", corn, "cornInterval", config.CRONINTERVAL(ctx), "注册成功")
	go RefreshSession(ctx)
}

func RefreshSession(ctx g.Ctx) {
	m := model.NewChatgptSession()
	result, err := cool.DBM(m).OrderAsc("updateTime").All()
	if err != nil {
		g.Log().Error(ctx, "RefreshSession", err)
		return
	}
	for _, v := range result {
		// 如果 v.mode ==1  则跳过
		g.Log().Info(ctx, "RefreshSession", v["email"], "start")
		if v["mode"].Int() == 1 {
			continue
		}

		getSessionUrl := config.CHATPROXY(ctx) + "/getsession"
		refreshCookie := gjson.New(v["officialSession"]).Get("refreshCookie").String()
		if refreshCookie == "" {
			continue
		}
		sessionVar := g.Client().SetHeader("authkey", config.AUTHKEY(ctx)).PostVar(ctx, getSessionUrl, g.Map{
			"username": v["email"],
			"password": v["password"],
			"authkey":  config.AUTHKEY(ctx),
			// "refreshCookie": refreshCookie,
		})
		g.Log().Info(ctx, "账号", v["email"], "sessionVar", sessionVar)
		sessionJson := gjson.New(sessionVar)
		if sessionJson.Get("accessToken").String() == "" {
			g.Log().Error(ctx, "RefreshSession", v["email"], "get session error", sessionJson)
			continue
		}
		_, err = cool.DBM(m).Where("email=?", v["email"]).Update(g.Map{
			"officialSession": sessionJson.String(),
		})
		if err != nil {
			g.Log().Error(ctx, "RefreshSession", err)
			continue
		}
		g.Log().Info(ctx, "RefreshSession", v["email"], "success")
		// 延时5分钟
		// time.Sleep(5 * time.Minute)
	}

}
