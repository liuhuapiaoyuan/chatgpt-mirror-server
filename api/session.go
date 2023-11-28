package api

import (
	backendapi "chatgpt-mirror-server/backend-api"
	"chatgpt-mirror-server/modules/chatgpt/model"
	"net/http"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Session(r *ghttp.Request) {
	ctx := r.GetCtx()
	userToken := r.Session.MustGet("userToken")
	record, expireTime, err := ChatgptSessionService.GetSessionByUserToken(ctx, userToken.String())
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	if record.IsEmpty() {
		g.Log().Error(ctx, "session is empty")
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	// officialSession := gjson.New(record["officialSession"].String())

	// getSessionUrl := config.CHATPROXY(ctx) + "/getsession"
	// refreshCookie := officialSession.Get("refreshCookie").String()
	// sessionVar := g.Client().SetHeader("authkey", config.AUTHKEY(ctx)).PostVar(ctx, getSessionUrl, g.Map{
	// 	"username":      record["email"].String(),
	// 	"password":      record["password"].String(),
	// 	"authkey":       config.AUTHKEY(ctx),
	// 	"refreshCookie": refreshCookie,
	// })
	sessionJson := gjson.New(record["officialSession"].String())
	if sessionJson.Get("accessToken").String() == "" {
		g.Log().Error(ctx, "get session error", sessionJson)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	cool.DBM(model.NewChatgptSession()).Where("email=?", record["email"].String()).Update(g.Map{
		"officialSession": sessionJson.String(),
	})
	backendapi.AccessTokenCache.Set(ctx, userToken.String(), sessionJson.Get("accessToken").String(), 10*24*time.Hour)
	sessionJson.Set("accessToken", userToken.String())
	sessionJson.Set("user.email", "admin@openai.com")
	sessionJson.Set("user.name", expireTime)
	sessionJson.Set("user.image", "/avatars.png")
	sessionJson.Set("user.picture", "/avatars.png")
	sessionJson.Remove("refreshCookie")

	r.Response.WriteJsonExit(sessionJson)
}
