package backendapi

import (
	"bytes"
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/model"
	"chatgpt-mirror-server/modules/chatgpt/service"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	ChatgptSessionService = service.NewChatgptSessionService()
)

func init() {
	s := g.Server()
	s.BindHandler("/backend-api/*any", ProxyAll)
	backendGroup := s.Group("/backend-api")
	backendGroup.POST("/accounts/data_export", NotFound) // 禁用导出
	backendGroup.POST("/payments/checkout", NotFound)    // 禁用支付
	// backendGroup.GET("/accounts/check/*any", accounts.Check)
	backendGroup.GET("/me", Me)
	backendGroup.GET("/conversations", Conversations)

}

// NotFound 404
func NotFound(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusNotFound)
}

// // 接受json字符串，处理完成后返回json字符串
// func _conversationListProcess(json string) {
// 	// 1. 转换成json对象

// 	// 2. 遍历每一个conversation
// 	// 3. 调用GetHistoryByUserIdAndConversationIds
// 	// 4. 将返回的结果添加到conversation中
// 	// 5. 返回json字符串
// }

func ProxyAll(r *ghttp.Request) {
	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := r.Header.Get("Authorization")[7:]

	userId, accessToken, err := ChatgptSessionService.GetAccessToken(ctx, userToken)

	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	UpStream := config.CHATPROXY(ctx)
	u, _ := url.Parse(UpStream)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		g.Log().Error(ctx, e)
		writer.WriteHeader(http.StatusBadGateway)
	}
	newreq := r.Request.Clone(ctx)
	newreq.URL.Host = u.Host
	newreq.URL.Scheme = u.Scheme
	newreq.Host = u.Host
	newreq.Header.Set("authkey", config.AUTHKEY(ctx))
	newreq.Header.Set("Authorization", "Bearer "+accessToken)

	// g.Dump(newreq.URL)
	cdnhost := config.CDNHOST(ctx)
	proxy.ModifyResponse = func(response *http.Response) error {
		path := response.Request.URL.Path
		// 如果path 以 ‘backend-api/files’ startwith开头
		isDownload := strings.HasPrefix(path, "/backend-api/files") && strings.HasSuffix(path, "download")
		isGizmosInfo := strings.HasPrefix(path, "/backend-api/gizmos/")
		isUpdateFile := strings.HasPrefix(path, "/backend-api/files") && r.Request.Method == "POST"
		isCreateConversation := strings.HasPrefix(path, "/backend-api/conversation/gen_title")
		if isCreateConversation {
			CreateConversation(ctx, userId, accessToken, r.UserAgent(), path)
		} else if isDownload || isUpdateFile || isGizmosInfo {

			g.Log().Info(ctx, "path", path)
			g.Log().Info(ctx, "path content-type",
				response.Header.Get("Content-Type"))
			originalBody, shouldReturn, returnValue := loadRespString(response)
			if shouldReturn {
				return returnValue
			}
			modifiedBody := strings.Replace(string(originalBody), "https://files.oaiusercontent.com", cdnhost, -1)
			// 将修改后的内容写回响应体
			response.Body = io.NopCloser(bytes.NewBufferString(modifiedBody))
			// 更新Content-Length
			response.ContentLength = int64(len(modifiedBody))
			response.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

			// 删除Content-Encoding头部
			response.Header.Del("Content-Encoding")
		}

		return nil
	}

	proxy.ServeHTTP(r.Response.Writer.RawWriter(), newreq)

}

// 创建信息，接受参数 conversationId
func CreateConversation(ctx g.Ctx, userId int, AccessToken string, userAgent string, conversationPath string) {
	// 提取 /backend-api/conversation/gen_title/{id}
	id := strings.Split(conversationPath, "/")[4]
	g.Log().Info(ctx, "提取出的ID", id)

	UpStream := config.CHATPROXY(ctx)
	// 请求后端接口
	res, err := g.Client().SetHeaderMap(map[string]string{
		"Authorization": "Bearer " + AccessToken,
		"User-Agent":    userAgent,
	}).Get(ctx, UpStream+"/backend-api/conversation/"+id)
	if err != nil {
		return
	}
	resStr := res.ReadAllString()
	resJson := gjson.New(resStr)
	g.Log().Info(ctx, resJson)

	history := model.NewChatgptHistory()
	history.UserId = userId
	history.Title = resJson.Get("title").String()
	history.CreateTime = resJson.Get("create_time").Time()
	history.UpdateTime = resJson.Get("update_time").Time()
	history.ConversationId = resJson.Get("conversation_id").String()

	conversationTemplateId := resJson.Get("conversation_template_id").String()
	if conversationTemplateId != "" {
		history.ConversationTemplate_id = conversationTemplateId
	}
	gizmo_id := resJson.Get("gizmo_id").String()
	if gizmo_id != "" {
		history.GizmoId = gizmo_id
	}
	cool.DBM(model.NewChatgptHistory()).InsertAndGetId(history)
}

func loadRespString(response *http.Response) ([]byte, bool, error) {
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":

		var err error
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, true, err
		}
		defer reader.Close()
	case "br":

		reader = io.NopCloser(brotli.NewReader(response.Body))
		defer reader.Close()
	default:
		reader = response.Body
	}

	originalBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, true, err
	}

	response.Body.Close()
	return originalBody, false, nil
}
