package arkose

import (
	"chatgpt-mirror-server/config"
	"net/http/httputil"
	"net/url"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	UpStream = config.ARKOSE_PROXY_URL
	proxy    *httputil.ReverseProxy
	Remote   *url.URL
)

func init() {
	remote, _ := url.Parse(UpStream)
	Remote = remote
	proxy = httputil.NewSingleHostReverseProxy(remote)

}

func ProxyArkose(r *ghttp.Request) {
	ctx := r.GetCtx()
	path := r.RequestURI
	// g.Log().Info(ctx, "ProxyArkose", path)
	// isAdmin := r.Session.MustGet("isAdmin").Bool()
	// if !isAdmin {
	// 	carid := r.Session.MustGet("carid").String()
	// 	_, err := utility.CheckCar(ctx, carid)
	// 	if err != nil {
	// 		g.Log().Error(ctx, err)
	// 		r.Response.Status = 401
	// 		r.Response.WriteJson(g.Map{
	// 			"detail": "Authentication credentials were not provided.",
	// 		})
	// 		return
	// 	}
	// }

	newreq := r.Request.Clone(ctx)
	newreq.URL.Host = Remote.Host
	newreq.URL.Scheme = Remote.Scheme
	newreq.Host = Remote.Host
	g.Log().Info(ctx, "ProxyArkose", path, "--->", newreq.URL.String())

	proxy.ServeHTTP(r.Response.Writer.RawWriter(), newreq)

}
