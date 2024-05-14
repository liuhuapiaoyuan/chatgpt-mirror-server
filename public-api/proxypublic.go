package publicapi

import (
	backendapi "chatgpt-mirror-server/backend-api"
	"chatgpt-mirror-server/config"
	"chatgpt-mirror-server/utility"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	u, _  = url.Parse(config.CHATPROXY)
	proxy = httputil.NewSingleHostReverseProxy(u)
)

func ProxyPublic(r *ghttp.Request) {
	ctx := r.GetCtx()
	userToken := ""
	Authorization := r.Header.Get("Authorization")
	if Authorization != "" {
		userToken = r.Header.Get("Authorization")[7:]
	}
	g.Log().Debug(ctx, "userToken", userToken)

	officialAccessToken := ""
	if userToken != "" {
		officialAccessToken = backendapi.AccessTokenCache.MustGet(ctx, userToken).String()
		if officialAccessToken == "" {
			record, _, err := backendapi.ChatgptSessionService.GetSessionByUserToken(ctx, userToken)
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
			backendapi.AccessTokenCache.Set(ctx, userToken, officialAccessToken, time.Minute)
		}
	}

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
	if officialAccessToken != "" {
		newreq.Header.Set("Authorization", "Bearer "+officialAccessToken)
	}
	// newreq.Header.Set("Cookie", "__Secure-next-auth.session-token="+carinfo.RefreshCookie)
	// // 去除header 中的 压缩
	// newreq.Header.Del("Accept-Encoding")
	proxy.ServeHTTP(r.Response.Writer.RawWriter(), newreq)
}
