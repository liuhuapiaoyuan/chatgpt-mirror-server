package publicapi

import (
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/utility"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ProxyPublic(r *ghttp.Request) {
	ctx := r.GetCtx()
	u, _ := url.Parse(config.CHATPROXY(ctx))
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		g.Log().Error(ctx, e)
		writer.WriteHeader(http.StatusBadGateway)
	}
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
	newreq := r.Request.Clone(ctx)
	newreq.URL.Host = u.Host
	newreq.URL.Scheme = u.Scheme
	newreq.Host = u.Host
	newreq.Header.Set("authkey", config.AUTHKEY(ctx))
	utility.HeaderModify(&newreq.Header)
	// newreq.Header.Set("Cookie", "__Secure-next-auth.session-token="+carinfo.RefreshCookie)
	// // 去除header 中的 压缩
	// newreq.Header.Del("Accept-Encoding")
	proxy.ServeHTTP(r.Response.Writer.RawWriter(), newreq)
}
