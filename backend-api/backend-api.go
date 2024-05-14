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
			AttachGPT4O(ctx, response)
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

// 处理models 增加gpt-4O
func AttachGPT4O(ctx g.Ctx, response *http.Response) error {
	// 提取 /backend-api/models
	originalBody, shouldReturn, returnValue := loadRespString(response)
	if shouldReturn {
		return returnValue
	}
	modifiedBody := string(originalBody)
	if !strings.Contains(modifiedBody, "gpt-4") {
		modifiedBody = `{
			"models": [
				{
					"slug": "text-davinci-002-render-sha",
					"max_tokens": 8191,
					"title": "Default (GPT-3.5)",
					"description": "我们最快的模型，非常适合大多数日常任务。",
					"tags": [
						"gpt3.5"
					],
					"capabilities": {},
					"product_features": {}
				},
				{
					"slug": "gpt-4o",
					"max_tokens": 8192,
					"title": "Scallion",
					"description": "浏览、高级数据分析和 DALL·E 现已集成到 GPT-4 中",
					"tags": [
						"confidential",
						"gpt4"
					],
					"capabilities": {},
					"product_features": {
						"attachments": {
							"type": "retrieval",
							"accepted_mime_types": [
								"text/x-java",
								"application/x-latext",
								"text/javascript",
								"text/x-script.python",
								"application/pdf",
								"text/html",
								"text/x-php",
								"text/x-c",
								"text/x-csharp",
								"application/vnd.openxmlformats-officedocument.presentationml.presentation",
								"text/x-tex",
								"application/json",
								"text/plain",
								"application/msword",
								"text/x-ruby",
								"text/markdown",
								"text/x-sh",
								"text/x-typescript",
								"text/x-c++",
								"application/vnd.openxmlformats-officedocument.wordprocessingml.document"
							],
							"image_mime_types": [
								"image/jpeg",
								"image/gif",
								"image/webp",
								"image/png"
							],
							"can_accept_all_mime_types": true
						}
					},
					"enabled_tools": [
						"tools",
						"tools2"
					]
				},
				{
					"slug": "auto",
					"max_tokens": 8192,
					"title": "Dynamic",
					"description": "使用合适的模型来满足我的请求",
					"tags": [
						"confidential",
						"gpt4"
					],
					"capabilities": {},
					"product_features": {
						"attachments": {
							"type": "retrieval",
							"accepted_mime_types": [
								"text/x-java",
								"application/x-latext",
								"text/javascript",
								"text/x-script.python",
								"application/pdf",
								"text/html",
								"text/x-php",
								"text/x-c",
								"text/x-csharp",
								"application/vnd.openxmlformats-officedocument.presentationml.presentation",
								"text/x-tex",
								"application/json",
								"text/plain",
								"application/msword",
								"text/x-ruby",
								"text/markdown",
								"text/x-sh",
								"text/x-typescript",
								"text/x-c++",
								"application/vnd.openxmlformats-officedocument.wordprocessingml.document"
							],
							"image_mime_types": [
								"image/jpeg",
								"image/gif",
								"image/webp",
								"image/png"
							],
							"can_accept_all_mime_types": true
						}
					},
					"enabled_tools": [
						"tools",
						"tools2"
					]
				}
			],
			"categories": [
				{
					"category": "gpt_3.5",
					"human_category_name": "GPT-3.5",
					"human_category_short_name": "3.5",
					"icon": "bolt",
					"icon_src": "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik0xMi41NjU1IDIuMTEwN0MxMy41Njg2IDAuOTIxNiAxNS40OTU2IDEuODU3NjEgMTUuMTgxIDMuMzgxMjJMMTQuMjI3NCA4LjAwMDAzSDE5LjkyMzlDMjEuMTk5OSA4LjAwMDAzIDIxLjg5MzIgOS40OTE5NyAyMS4wNzA0IDEwLjQ2NzNMMTEuNDM0MyAyMS44ODk0QzEwLjQzMTEgMjMuMDc4NSA4LjUwNDI0IDIyLjE0MjUgOC44MTg3OSAyMC42MTg5TDkuNzcyMzYgMTZINC4wNzU4OEMyLjc5OTg3IDE2IDIuMTA2NTggMTQuNTA4MSAyLjkyOTM4IDEzLjUzMjhMMTIuNTY1NSAyLjExMDdaIiBmaWxsPSIjMjgyODI4Ii8+Cjwvc3ZnPgo=",
					"subscription_level": "free",
					"default_model": "text-davinci-002-render-sha",
					"code_interpreter_model": "text-davinci-002-render-sha-code-interpreter",
					"plugins_model": "text-davinci-002-render-sha-plugins",
					"color": "#47C761",
					"short_explainer": "非常适合用于日常任务",
					"tagline": "最快速"
				},
				{
					"category": "AG8PqS2q",
					"human_category_name": "GPT-4o",
					"human_category_short_name": "4o",
					"icon": "stars",
					"icon_src": "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTE5LjkxOTQgMC44OTY4ODhDMTkuODkzOCAwLjY3MDk5NCAxOS43MDI4IDAuNTAwMjMgMTkuNDc1NSAwLjVDMTkuMjQ4MSAwLjQ5OTc3IDE5LjA1NjcgMC42NzAxMzEgMTkuMDMwOCAwLjg5NTk4QzE4LjkwOTQgMS45NDc0IDE4LjU5NzIgMi42Njg3MiAxOC4xMDggMy4xNTc5NkMxNy42MTg3IDMuNjQ3MjEgMTYuODk3NCAzLjk1OTQ1IDE1Ljg0NTkgNC4wODA3MkMxNS42MjAxIDQuMTA2NzcgMTUuNDQ5OCA0LjI5ODEyIDE1LjQ1IDQuNTI1NDZDMTUuNDUwMiA0Ljc1MjggMTUuNjIxIDQuOTQzOCAxNS44NDY5IDQuOTY5MzhDMTYuODgwNSA1LjA4NjQ2IDE3LjYxODMgNS4zOTg2NCAxOC4xMiA1Ljg5MTI1QzE4LjYxODkgNi4zODExOCAxOC45MzcgNy4xMDE0NyAxOS4wMjk1IDguMTQyMzVDMTkuMDUwMSA4LjM3MzI5IDE5LjI0MzYgOC41NTAyNiAxOS40NzU1IDguNTVDMTkuNzA3MyA4LjU0OTc0IDE5LjkwMDYgOC4zNzIzNCAxOS45MjA1IDguMTQxMzVDMjAuMDA5MiA3LjExODA1IDIwLjMyNyA2LjM4MTU0IDIwLjgyOTIgNS44NzkyOUMyMS4zMzE2IDUuMzc3MDUgMjIuMDY4IDUuMDU5MTMgMjMuMDkxMyA0Ljk3MDU2QzIzLjMyMjMgNC45NTA1NiAyMy40OTk4IDQuNzU3MzYgMjMuNSA0LjUyNTUxQzIzLjUwMDIgNC4yOTM2NSAyMy4zMjMyIDQuMTAwMDUgMjMuMDkyMyA0LjA3OTU0QzIyLjA1MTUgMy45ODcwOCAyMS4zMzEyIDMuNjY4OTEgMjAuODQxMiAzLjE2OTk2QzIwLjM0ODcgMi42NjgyNyAyMC4wMzY0IDEuOTMwNTMgMTkuOTE5NCAwLjg5Njg4OFoiIGZpbGw9IiMyODI4MjgiLz4KPHBhdGggZD0iTTExLjk5MjYgMy44MjA1N0MxMS45MjY5IDMuMjM5NjkgMTEuNDM1OCAyLjgwMDYgMTAuODUxMiAyLjhDMTAuMjY2NiAyLjc5OTQgOS43NzQ1MiAzLjIzNzQ4IDkuNzA3NTkgMy44MTgyM0M5LjM5NTczIDYuNTIxOTEgOC41OTI4MyA4LjM3NjY4IDcuMzM0NzYgOS42MzQ3NkM2LjA3NjcgMTAuODkyOCA0LjIyMTkxIDExLjY5NTcgMS41MTgyMyAxMi4wMDc2QzAuOTM3NDg0IDEyLjA3NDUgMC40OTk0MDMgMTIuNTY2NiAwLjUwMDAwMSAxMy4xNTEyQzAuNTAwNTk5IDEzLjczNTggMC45Mzk2OTIgMTQuMjI2OSAxLjUyMDU3IDE0LjI5MjZDNC4xNzg1MiAxNC41OTM3IDYuMDc1NTUgMTUuMzk2NSA3LjM2NTU5IDE2LjY2MzNDOC42NDg2MiAxNy45MjMxIDkuNDY2NzYgMTkuNzc1MSA5LjcwNDQ4IDIyLjQ1MThDOS43NTcyNyAyMy4wNDU2IDEwLjI1NTEgMjMuNTAwNyAxMC44NTEzIDIzLjVDMTEuNDQ3NSAyMy40OTkzIDExLjk0NDMgMjMuMDQzMSAxMS45OTU3IDIyLjQ0OTFDMTIuMjIzNCAxOS44MTc4IDEzLjA0MSAxNy45MjQgMTQuMzMyNCAxNi42MzI0QzE1LjYyNCAxNS4zNDEgMTcuNTE3OCAxNC41MjM0IDIwLjE0OTEgMTQuMjk1N0MyMC43NDMxIDE0LjI0NDMgMjEuMTk5MyAxMy43NDc1IDIxLjIgMTMuMTUxM0MyMS4yMDA3IDEyLjU1NTEgMjAuNzQ1NiAxMi4wNTczIDIwLjE1MTggMTIuMDA0NUMxNy40NzUxIDExLjc2NjggMTUuNjIzMSAxMC45NDg3IDE0LjM2MzMgOS42NjU1OUMxMy4wOTY1IDguMzc1NTUgMTIuMjkzNyA2LjQ3ODUyIDExLjk5MjYgMy44MjA1N1oiIGZpbGw9IiMyODI4MjgiLz4KPC9zdmc+Cg==",
					"subscription_level": "plus",
					"default_model": "gpt-4o",
					"color": "#5B41F9",
					"short_explainer": "Newest and most advanced model",
					"tagline": "智能且快速"
				},
				{
					"category": "auto",
					"human_category_name": "Dynamic",
					"human_category_short_name": "",
					"icon": "connected",
					"icon_src": "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgZmlsbD0ibm9uZSIgdmlld0JveD0iMCAwIDI0IDI0Ij48cGF0aCBmaWxsPSJjdXJyZW50Q29sb3IiIGZpbGwtcnVsZT0iZXZlbm9kZCIgZD0iTTEyIDcuNDJhMjIgMjIgMCAwIDAtMi40NTMgMi4xMjdBMjIgMjIgMCAwIDAgNy40MiAxMmEyMiAyMiAwIDAgMCAyLjEyNyAyLjQ1M2MuODA3LjgwOCAxLjYzNiAxLjUyIDIuNDUzIDIuMTI4YTIyIDIyIDAgMCAwIDIuNDUzLTIuMTI4QTIyIDIyIDAgMCAwIDE2LjU4IDEyYTIyIDIyIDAgMCAwLTIuMTI3LTIuNDUzQTIyIDIyIDAgMCAwIDEyIDcuNDJtMS43NTEtMS4xNTRhMjUgMjUgMCAwIDEgMi4xMDQgMS44OCAyNSAyNSAwIDAgMSAxLjg4IDIuMTAzYy4zMTYtLjU1LjU3Ni0xLjA4NS43NzktMS41OS4zNS0uODc4LjUwNy0xLjYyNS41MDMtMi4yMDYtLjAwMy0uNTc0LS4xNi0uOTEzLS4zNTgtMS4xMTEtLjE5OS0uMTk5LS41MzctLjM1Ni0xLjExMi0uMzYtLjU4LS4wMDMtMS4zMjguMTUzLTIuMjA1LjUwNC0uNTA2LjIwMy0xLjA0LjQ2NC0xLjU5Ljc4Wm0zLjk4MyA3LjQ4NWEyNSAyNSAwIDAgMS0xLjg4IDIuMTA0IDI1IDI1IDAgMCAxLTIuMTAzIDEuODggMTMgMTMgMCAwIDAgMS41OS43NzljLjg3OC4zNSAxLjYyNS41MDcgMi4yMDYuNTAzLjU3NC0uMDAzLjkxMy0uMTYgMS4xMTEtLjM1OC4xOTktLjE5OS4zNTYtLjUzOC4zNi0xLjExMi4wMDMtLjU4LS4xNTQtMS4zMjgtLjUwNC0yLjIwNWExMyAxMyAwIDAgMC0uNzgtMS41OVpNMTIgMTguOTljLjg5LjU3IDEuNzY4IDEuMDMgMi42MDUgMS4zNjQgMS4wMjYuNDEgMi4wMzYuNjUyIDIuOTU1LjY0Ni45MjUtLjAwNiAxLjgyOC0uMjY3IDIuNS0uOTQuNjczLS42NzIuOTM0LTEuNTc1Ljk0LTIuNS4wMDYtLjkxOS0uMjM2LTEuOTI5LS42NDYtMi45NTRBMTUuNyAxNS43IDAgMCAwIDE4Ljk5IDEyYTE1LjYgMTUuNiAwIDAgMCAxLjM2NC0yLjYwNmMuNDEtMS4wMjUuNjUyLTIuMDM1LjY0Ni0yLjk1NC0uMDA2LS45MjUtLjI2Ny0xLjgyOC0uOTQtMi41LS42NzItLjY3My0xLjU3NS0uOTM0LTIuNS0uOTQtLjkxOS0uMDA2LTEuOTI5LjIzNS0yLjk1NC42NDYtLjgzOC4zMzUtMS43MTYuNzk1LTIuNjA2IDEuMzY0YTE1LjcgMTUuNyAwIDAgMC0yLjYwNi0xLjM2NEM4LjM3IDMuMjM2IDcuMzYgMi45OTQgNi40NCAzYy0uOTI1LjAwNi0xLjgyOC4yNjctMi41Ljk0LS42NzMuNjcyLS45MzQgMS41NzUtLjk0IDIuNS0uMDA2LjkxOS4yMzUgMS45MjkuNjQ2IDIuOTU1QTE1LjcgMTUuNyAwIDAgMCA1LjAxIDEyYy0uNTcuODktMS4wMyAxLjc2OC0xLjM2NCAyLjYwNS0uNDEgMS4wMjYtLjY1MiAyLjAzNi0uNjQ2IDIuOTU1LjAwNi45MjUuMjY3IDEuODI4Ljk0IDIuNS42NzIuNjczIDEuNTc1LjkzNCAyLjUuOTQuOTIuMDA2IDEuOTMtLjIzNSAyLjk1NS0uNjQ2QTE1LjcgMTUuNyAwIDAgMCAxMiAxOC45OW0tMS43NTEtMS4yNTVhMjUgMjUgMCAwIDEtMi4xMDQtMS44OCAyNSAyNSAwIDAgMS0xLjg4LTIuMTA0Yy0uMzE1LjU1LS41NzYgMS4wODUtLjc3OSAxLjU5LS4zNS44NzgtLjUwNyAxLjYyNS0uNTAzIDIuMjA2LjAwMy41NzQuMTYuOTEzLjM1OSAxLjExMS4xOTguMTk5LjUzNy4zNTYgMS4xMTEuMzYuNTguMDAzIDEuMzI4LS4xNTMgMi4yMDUtLjUwNC41MDYtLjIwMyAxLjA0LS40NjMgMS41OS0uNzhabS0zLjk4My03LjQ4NmEyNSAyNSAwIDAgMSAxLjg4LTIuMTA0IDI1IDI1IDAgMCAxIDIuMTAzLTEuODggMTMgMTMgMCAwIDAtMS41OS0uNzc5Yy0uODc4LS4zNS0xLjYyNS0uNTA3LTIuMjA2LS41MDMtLjU3NC4wMDMtLjkxMy4xNi0xLjExMS4zNTktLjE5OS4xOTgtLjM1Ni41MzctLjM2IDEuMTExLS4wMDMuNTguMTUzIDEuMzI4LjUwNCAyLjIwNS4yMDMuNTA2LjQ2NCAxLjA0Ljc4IDEuNTlaIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiPjwvcGF0aD48L3N2Zz4=",
					"subscription_level": "free",
					"default_model": "auto",
					"plugins_model": "gpt-4-plugins",
					"color": "",
					"short_explainer": "为改善速度和智能程度而进行了优化。",
					"tagline": ""
				}
			]
		}`
	}
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
