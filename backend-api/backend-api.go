package backendapi

import (
	"bytes"
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/model"
	"chatgpt-mirror-server/modules/chatgpt/service"
	"chatgpt-mirror-server/utility"
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
	"github.com/gogf/gf/v2/os/gcache"
)

var (
	ChatgptSessionService = service.NewChatgptSessionService()
	AccessTokenCache      = gcache.New()
)

func init() {
	s := g.Server()
	s.BindHandler("/backend-api/*any", ProxyAll)
	// s.BindHandler("/public-api/*any", ProxyAll)
	//s.BindHandler("/_next/data/*any", NextDataGptsFixed)
	backendGroup := s.Group("/backend-api")
	backendGroup.POST("/accounts/data_export", NotFound) // 禁用导出
	backendGroup.POST("/payments/checkout", NotFound)    // 禁用支付
	backendGroup.ALL("/accounts/*/invites", NotFound)    // 禁用邀请
	backendGroup.ALL("/accounts/*/users/*", NotFound)    // 成员
	backendGroup.ALL("/accounts/transfer", NotFound)     // 转移
	backendGroup.ALL("/accounts/logout_all", NotFound)   // 登出
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

// 代理请求
func ProxyRequestGet(path string, r *ghttp.Request) (resStr string, err error) {
	ctx := r.GetCtx()
	userToken := ""
	Authorization := r.Header.Get("Authorization")
	if Authorization != "" {
		userToken = r.Header.Get("Authorization")[7:]
	}
	_, _, accessToken, err := ChatgptSessionService.GetAccessToken(ctx, userToken)

	UpStream := config.CHATPROXY
	if err != nil {
		// 处理错误
		panic(err)
	}

	// 设置HTTP Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if config.Ja3Proxy != nil {
		transport = &http.Transport{
			Proxy: http.ProxyURL(config.Ja3Proxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ForceAttemptHTTP2: true,
		}
	}
	// 创建HTTP客户端
	client := &http.Client{Transport: transport}

	// 设置请求头
	req, err := http.NewRequestWithContext(ctx, "GET", UpStream+path, nil)
	if err != nil {
		// 处理错误
		panic(err)
	}
	// 遍历原始请求的Header，并将它们复制到另一个请求的Header中
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("authkey", config.AUTHKEY(ctx))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Origin", "https://chat.openai.com/chat")
	req.Header.Set("Referer", "https://chat.openai.com/")
	utility.HeaderModify(&req.Header)
	resp, err := client.Do(req)
	if err != nil {
		// 处理错误
		g.Log().Error(ctx, err)
		panic(err)
	}
	defer resp.Body.Close()
	originalBody, shouldReturn, err := loadRespString(resp)
	if err != nil || shouldReturn {
		return "", err
	}
	return string(originalBody), nil

}

func ProxyAll(r *ghttp.Request) {

	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := ""
	Authorization := r.Header.Get("Authorization")
	if Authorization != "" {
		userToken = r.Header.Get("Authorization")[7:]
	}
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
	WsUpStream := config.WS_SERVICE(ctx)
	// g.Log().Info(ctx, "ProxyBackendApi:", path)
	proxy := &httputil.ReverseProxy{}
	UpStream := config.CHATPROXY

	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2: true,
	}
	if config.Ja3Proxy != nil {
		proxy.Transport = &http.Transport{
			Proxy: http.ProxyURL(config.Ja3Proxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ForceAttemptHTTP2: true,
		}
	}
	OPENAI, err := url.Parse(UpStream)
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatus(http.StatusServiceUnavailable)
		return
	}
	proxy.Rewrite = func(proxyRequest *httputil.ProxyRequest) {
		proxyRequest.SetURL(OPENAI)
	}

	header := r.Request.Header
	header.Set("Origin", "https://chat.openai.com")
	header.Set("Referer", "https://chat.openai.com/")
	// header.Del("Cookie")
	header.Del("Accept-Encoding")
	if accessToken != "" {
		header.Set("Authorization", "Bearer "+accessToken)
	}
	utility.HeaderModify(&r.Request.Header)

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
		if strings.Contains(response.Header.Get("Content-Type"), "json") {
			isStream = false
		}
		if isCreateConversation {
			CreateConversation(r, userId, chatgptId, userToken, r.UserAgent(), path)
		} else if isLoadModels {
			//AttachGPT4Mobile(ctx, response)
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
				re := regexp.MustCompile(`wss://([^/]+)/client`)

				// 在URL中搜索匹配的部分
				matches := re.FindStringSubmatch(bodyStr)

				if len(matches) > 1 {
					modifiedBody = strings.Replace(modifiedBody, "wss://"+matches[1]+"/client/hubs/conversations?", WsUpStream+"/client/hubs/conversations?host="+matches[1]+"&", -1)
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

	proxy.ServeHTTP(r.Response.RawWriter(), r.Request)

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
func CreateConversation(r *ghttp.Request, userId int, chatgptId int, userToken string, userAgent string, conversationPath string) {
	id := strings.Split(conversationPath, "/")[4]
	r.Request.Header.Set("Authorization", "Bearer "+userToken)
	g.Log().Info(r.GetCtx(), "准备开始创建话题CreateConversation", id)
	resStr, err := ProxyRequestGet("/backend-api/conversation/"+id, r)
	if err != nil {
		return
	}
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
