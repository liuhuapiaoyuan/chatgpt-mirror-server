package api

import (
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/model"
	"chatgpt-mirror-server/modules/chatgpt/service"
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

	// 如果record mode==1
	if record["mode"].Int() == 1 {
		// 如果没有officialSession，就去获取
		if record["officialSession"].String() == "" || record["officialSession"].String() == "null" {
			g.Log().Error(ctx, "手工模式，没有token，无法登录")

			r.Response.WriteStatus(http.StatusUnauthorized)
			return
		}
		sessionJson := gjson.New("{}")
		sessionJson.Set("accessToken", record["officialSession"].String())
		service.AccessTokenCache.Set(ctx, userToken.String(), sessionJson.Get("accessToken").String(), 10*24*time.Hour)
		sessionJson.Set("accessToken", userToken.String())
		sessionJson.Set("user.email", "admin@openai.com")
		sessionJson.Set("user.name", expireTime)
		sessionJson.Set("user.image", "/avatars.png")
		sessionJson.Set("user.picture", "/avatars.png")
		sessionJson.Remove("refreshCookie")

		r.Response.WriteJsonExit(sessionJson)
	} else {
		sessionJson := gjson.New(record["officialSession"].String())
		getSessionUrl := config.CHATPROXY(ctx) + "/auth/token"
		expires := sessionJson.Get("expires").Time()

		// 判断是否过期
		if expires.Before(time.Now()) {
			g.Log().Info(ctx, "session 过期，重新获取")
			sessionVar := g.Client().SetHeader("authkey", config.AUTHKEY(ctx)).PostVar(ctx, getSessionUrl, g.Map{
				"username": record["email"].String(),
				"password": record["password"].String(),
				"authkey":  config.AUTHKEY(ctx),
			})
			sessionJson = gjson.New(sessionVar)
			if sessionJson.Get("accessToken").String() == "" {
				g.Log().Error(ctx, "get session error", sessionJson)
				r.Response.WriteStatus(http.StatusUnauthorized)
				return
			}
			cool.DBM(model.NewChatgptSession()).Where("email=?", record["email"].String()).Update(g.Map{
				"officialSession": sessionJson.String(),
			})
		}

		service.AccessTokenCache.Set(ctx, userToken.String(), sessionJson.Get("accessToken").String(), 10*24*time.Hour)
		sessionJson.Set("accessToken", userToken.String())
		sessionJson.Set("user.email", "admin@openai.com")
		sessionJson.Set("user.name", expireTime)
		sessionJson.Set("user.image", "/avatars.png")
		sessionJson.Set("user.picture", "/avatars.png")
		sessionJson.Remove("refreshCookie")

		r.Response.WriteJsonExit(sessionJson)

	}

}
