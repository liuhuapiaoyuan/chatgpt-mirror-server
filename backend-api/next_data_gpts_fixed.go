package backendapi

import (
	"chatgpt-mirror-server/config"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func NextDataGptsFixed(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 如果请求包含gizmoId
	gizmoId := r.Request.URL.Query().Get("gizmoId")
	if gizmoId != "" {
		UpStream := config.CHATPROXY
		u, _ := url.Parse(UpStream)
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, ForceAttemptHTTP2: true,
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
		proxy.ServeHTTP(r.Response.Writer.RawWriter(), newreq)
		return

	}

	// r.Response.Header().Set("X-Client-Source", "explorer")
	// r.Response.Header().Set("X-Middleware-Skip", "1")
	// r.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	// r.Response.Write(`{}`)
}
