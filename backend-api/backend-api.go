package backendapi

import (
	"bytes"
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/modules/chatgpt/service"
	"chatgpt-mirror-server/utility"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
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
	backendGroup := s.Group("/backend-api")
	backendGroup.POST("/accounts/data_export", NotFound) // 禁用导出
	backendGroup.POST("/payments/checkout", NotFound)    // 禁用支付
	// backendGroup.GET("/accounts/check/*any", accounts.Check)
	backendGroup.GET("/me", Me)

}

// NotFound 404
func NotFound(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusNotFound)
}

func ProxyAll(r *ghttp.Request) {
	ctx := r.GetCtx()
	// 获取header中的token Authorization: Bearer xxx 去掉Bearer
	userToken := r.Header.Get("Authorization")[7:]

	officialAccessToken := AccessTokenCache.MustGet(ctx, userToken).String()
	if officialAccessToken == "" {
		record, _, err := ChatgptSessionService.GetSessionByUserToken(ctx, userToken)
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
		officialSession := record["officialSession"].String()
		if officialSession == "" {
			r.Response.WriteStatus(http.StatusUnauthorized)
			return
		}
		officialAccessToken = utility.AccessTokenFormSession(officialSession)
		AccessTokenCache.Set(ctx, userToken, officialAccessToken, time.Minute)
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
	newreq.Header.Set("Authorization", "Bearer "+officialAccessToken)

	// g.Dump(newreq.URL)
	cdnhost := config.CDNHOST(ctx)
	proxy.ModifyResponse = func(response *http.Response) error {
		path := response.Request.URL.Path
		// 如果path 以 ‘backend-api/files’ startwith开头
		isDownload := strings.HasPrefix(path, "/backend-api/files") && strings.HasSuffix(path, "download")
		isGizmosInfo := strings.HasPrefix(path, "/backend-api/gizmos/")
		if isDownload || isGizmosInfo {
			g.Log().Info(ctx, "path", path)
			g.Log().Info(ctx, "path content-type",
				response.Header.Get("Content-Type"))
			var reader io.ReadCloser
			switch response.Header.Get("Content-Encoding") {
			case "gzip":
				// 处理 gzip 压缩
				var err error
				reader, err = gzip.NewReader(response.Body)
				if err != nil {
					return err
				}
				defer reader.Close()
			case "br":
				// 处理 Brotli 压缩
				reader = io.NopCloser(brotli.NewReader(response.Body))
				defer reader.Close()
			default:
				reader = response.Body
			}
			// 读取原始响应体
			originalBody, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			// 关闭原始响应体
			response.Body.Close()
			// 替换字符串 "https://proxy.ggss.club/file"

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
