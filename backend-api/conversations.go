package backendapi

import (
	"chatgpt-mirror-server/modules/chatgpt/model"
	"net/http"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// 获得聊天记录
func Conversations(r *ghttp.Request) {
	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := r.Header.Get("Authorization")[7:]
	user, err2 := cool.DBM(model.NewChatgptUser()).Where("userToken", userToken).One()
	if err2 != nil {
		g.Log().Error(ctx, err2)
		return
	}
	userId := user["id"].Int()

	// 获得分页参数 s?offset=0&limit=28&order=updated
	offset := r.GetQuery("offset").Int()
	limit := r.GetQuery("limit").Int()
	order := "updateTime desc"

	list, err := cool.DBM(model.NewChatgptHistory()).Where("user_id=?", userId).Order(order).Limit(offset, limit).All()
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	total, err := cool.DBM(model.NewChatgptHistory()).Where("user_id=?", userId).Order(order).Count()
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	// 遍历查询结果
	for _, record := range list {
		record["id"] = record["conversation_id"]
		if createTime, ok := record["createTime"]; ok {
			record["create_time"] = createTime
		}
		if updateTime, ok := record["updateTime"]; ok {
			record["update_time"] = updateTime
		}
	}

	resJson := gjson.New("{}")
	resJson.Set("items", list)
	resJson.Set("offset", offset)
	resJson.Set("limit", limit)
	resJson.Set("total", total)
	resJson.Set("has_missing_conversations", false)
	// 遍历 resJson['items'] 修改时间戳

	r.Response.WriteJson(resJson)
}
