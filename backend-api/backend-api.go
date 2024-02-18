package backendapi

import (
	"bytes"
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/model"
	"chatgpt-mirror-server/modules/chatgpt/service"
	"compress/gzip"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
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
	// s.BindHandler("/public-api/*any", ProxyAll)
	//	s.BindHandler("/_next/data/*any", NextDataGptsFixed)
	backendGroup := s.Group("/backend-api")
	backendGroup.POST("/accounts/data_export", NotFound) // 禁用导出
	backendGroup.POST("/payments/checkout", NotFound)    // 禁用支付
	backendGroup.ALL("/accounts/*/invites", NotFound)    // 禁用邀请
	// backendGroup.GET("/accounts/check/*any", accounts.Check)
	backendGroup.GET("/me", Me)
	backendGroup.GET("/conversations", Conversations)
	backendGroup.GET("/prompt_library", Prompt_library)
	backendGroup.GET("/upgrade_invites", NextDataGptsFixed)

}

// NotFound 404
func NotFound(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusNotFound)
}

func ProxyAll(r *ghttp.Request) {

	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := ""
	Authorization := r.Header.Get("Authorization")
	if Authorization != "" {
		userToken = r.Header.Get("Authorization")[7:]
	}
	g.Log().Debug(ctx, "userToken", userToken)

	isStream := strings.Contains(r.Header.Get("accept"), "text/event-stream")

	// 获得当前的请求域名
	// g.Log().Debug(ctx, "ProxyAll", r.URL.Path, r.Header.Get("accept"), isStream)

	domain := r.Host
	userId, chatgptId, accessToken, err := ChatgptSessionService.GetAccessToken(ctx, userToken)

	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	UpStream := config.CHATPROXY(ctx)
	u, _ := url.Parse(UpStream)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		g.Log().Error(ctx, e)
		writer.WriteHeader(http.StatusBadGateway)
	}
	newreq := r.Request.Clone(ctx)
	newreq.URL.Host = u.Host
	newreq.URL.Scheme = u.Scheme
	newreq.Host = u.Host
	newreq.Header.Set("authkey", config.AUTHKEY(ctx))

	newreq.Header.Set("Host", "chat.openai.com")
	newreq.Header.Set("Origin", "https://chat.openai.com/chat")
	if accessToken != "" {
		newreq.Header.Set("Authorization", "Bearer "+accessToken)
	}

	// g.Dump(newreq.URL)
	cdnhost := config.CDNHOST(ctx)
	proxy.ModifyResponse = func(response *http.Response) error {
		path := response.Request.URL.Path
		// 如果path 以 ‘backend-api/files’ startwith开头
		isCreateConversation := strings.HasPrefix(path, "/backend-api/conversation/gen_title")
		isShare := strings.HasPrefix(path, "/backend-api/share/creat")
		isLoadModels := strings.HasPrefix(path, "/backend-api/models")
		if strings.HasPrefix(path, "/backend-api/conversation") {
			// log content-type
			g.Log().Debug(ctx, "content-type", response.Header.Get("Content-Type"))
		}

		// 判断response的Cotnent-Type是否是json
		if response.Header.Get("Content-Type") == "application/json" {
			isStream = false
		}
		if isCreateConversation {
			CreateConversation(ctx, userId, chatgptId, accessToken, r.UserAgent(), path)
		} else if isLoadModels {
			AttachGPT4Mobile(ctx, response)
		} else if isShare {
			originalBody, shouldReturn, returnValue := loadRespString(response)
			if shouldReturn {
				return returnValue
			}
			modifiedBody := strings.Replace(string(originalBody), "chat.openai.com/share", domain+"/share", -1)
			// 将修改后的内容写回响应体
			response.Body = io.NopCloser(bytes.NewBufferString(modifiedBody))
			// 更新Content-Length
			response.ContentLength = int64(len(modifiedBody))
			response.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
			// 删除Content-Encoding头部
			response.Header.Del("Content-Encoding")
		} else if !isStream {
			originalBody, shouldReturn, returnValue := loadRespString(response)
			if shouldReturn {
				return returnValue
			}
			bodyStr := string(originalBody)
			// 判断字符串是否包含 wss://
			modifiedBody := strings.Replace(bodyStr, "https://files.oaiusercontent.com", cdnhost, -1)

			if strings.Contains(bodyStr, "wss://") {
				// 使用正则 wss://中间的域名/client/hubs/conversations  把域名替换成本地域名

				// modifiedBody = strings.Replace(string(modifiedBody), "wss://chatgpt-async-webps-prod-southcentralus-22.webpubsub.azure.com",
				// "wss://host.docker.internal:7999", -1)

				re := regexp.MustCompile(`wss://([^/]+)/client`)

				// 在URL中搜索匹配的部分
				matches := re.FindStringSubmatch(bodyStr)

				if len(matches) > 1 {
					modifiedBody = strings.Replace(modifiedBody, "wss://"+matches[1]+"/client/hubs/conversations?", "ws://127.0.0.1:7999/client/hubs/conversations?host="+matches[1]+"&", -1)
					g.Log().Debug(ctx, "wss替换成ws", matches[1])
				} else {

					g.Log().Debug(ctx, "wss替换成ws失败", bodyStr)
				}
			}

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

// 处理models 增加gpt-4-mobile
func AttachGPT4Mobile(ctx g.Ctx, response *http.Response) error {
	// 提取 /backend-api/models
	originalBody, shouldReturn, returnValue := loadRespString(response)
	if shouldReturn {
		return returnValue
	}
	modifiedBody := string(originalBody)
	// if strings.Contains(modifiedBody, "gpt-4") {
	// }
	resJson := gjson.New(modifiedBody)
	models := resJson.Get("models").Array()
	newObject := gjson.New(`{"capabilities":{},"description":"Browsing, Advanced Data Analysis, and DALL·E are now built into GPT-4","enabled_tools":["tools","tools2"],"max_tokens":32767,"product_features":{"attachments":{"accepted_mime_types":["text/x-csharp","application/vnd.openxmlformats-officedocument.wordprocessingml.document","text/x-tex","text/x-typescript","text/plain","text/x-ruby","application/msword","text/x-php","text/x-c++","text/markdown","application/x-latext","text/x-c","text/javascript","text/html","application/vnd.openxmlformats-officedocument.presentationml.presentation","application/json","text/x-java","application/pdf","text/x-script.python","text/x-sh"],"can_accept_all_mime_types":true,"image_mime_types":["image/jpeg","image/webp","image/gif","image/png"],"type":"retrieval"}},"slug":"gpt-4-mobile","tags":["confidential","gpt4","plus"],"title":"GPT4 (Mobile)"}`)
	models = append(models, newObject)
	resJson.Set("models", models)

	categories := resJson.Get("categories").Array()

	categories = append(categories, gjson.New(`{
		"category": "gpt_4",
		"human_category_name": "GPT-4",
		"subscription_level": "plus",
		"default_model": "gpt-4",
		"browsing_model": "gpt-4-browsing",
		"code_interpreter_model": "gpt-4-code-interpreter",
		"plugins_model": "gpt-4-plugins"
	}`))
	resJson.Set("categories", categories)
	modifiedBody = resJson.String()
	// 将修改后的内容写回响应体
	response.Body = io.NopCloser(bytes.NewBufferString(modifiedBody))
	// 更新Content-Length
	response.ContentLength = int64(len(modifiedBody))
	response.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
	// 删除Content-Encoding头部
	response.Header.Del("Content-Encoding")
	return nil
}

// 创建信息，接受参数 conversationId
func CreateConversation(ctx g.Ctx, userId int, chatgptId int, AccessToken string, userAgent string, conversationPath string) {
	// 提取 /backend-api/conversation/gen_title/{id}
	id := strings.Split(conversationPath, "/")[4]

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

	history := model.NewChatgptHistory()
	history.UserId = userId
	history.Title = resJson.Get("title").String()
	history.CreateTime = resJson.Get("create_time").Time()
	history.UpdateTime = resJson.Get("update_time").Time()
	history.ConversationId = resJson.Get("conversation_id").String()
	// 账号ID
	history.ChatgptId = chatgptId

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

// 处理删除消息
func RemoveCreateConversation(ctx g.Ctx, response *http.Response, conversationPath string) {
	g.Log().Debug(ctx, "RemoveCreateConversation", conversationPath)
	id := strings.Split(conversationPath, "/")[3]
	g.Log().Debug(ctx, "提取出的ID", id)

	originalBody, shouldReturn, err := loadRespString(response)
	if err != nil || shouldReturn {
		return
	}
	resJson := gjson.New(string(originalBody))
	g.Log().Debug(ctx, "conversation will removed", resJson)
	if resJson.Get("success").Bool() {
		cool.DBM(model.NewChatgptHistory()).Where("conversation_id", id).Delete()
	}
	modifiedBody := string(originalBody)
	response.Body = io.NopCloser(bytes.NewBufferString(modifiedBody))
	// 更新Content-Length
	response.ContentLength = int64(len(modifiedBody))
	response.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
	// 删除Content-Encoding头部
	response.Header.Del("Content-Encoding")

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
