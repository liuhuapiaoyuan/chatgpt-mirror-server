package backendapi

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Me(r *ghttp.Request) {
	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := r.Header.Get("Authorization")[7:]
	record, expireTime, err := ChatgptSessionService.GetSessionByUserToken(ctx, userToken)
	g.Log().Info(ctx, "expireTime", expireTime)
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	resStr, err := ProxyRequestGet("/backend-api/me", r)
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	resJson := gjson.New(resStr)
	resJson.Set("email", "admin@closeai.com")
	resJson.Set("name", record["user_username"].String()+expireTime)
	resJson.Set("picture", "/avatars.png")
	resJson.Set("phone_number", "+1911011")
	resJson.Set("orgs.data.0.description", "closeai")
	// resJson.Dump()
	r.Response.WriteJson(resJson)
}
