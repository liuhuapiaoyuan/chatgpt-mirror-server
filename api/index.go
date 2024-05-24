
package api

import (
	"chatgpt-mirror-server/config"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Index(r *ghttp.Request) {

	ctx := r.GetCtx()
	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		// r.Response.Writer.Write([]byte("Hello XyHelper"))
		return
	}
	model := r.Get("model").String()

	propsJson := gjson.New(Props)
	if model != "" {
		propsJson.Set("query.model", model)
	}
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)

	r.Response.WriteTpl(config.CacheBuildId+"/chat.html", g.Map{
		"props":       propsJson,
		"arkoseUrl":   config.ArkoseUrl,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(ctx),
	})
}
func C(r *ghttp.Request) {
	ctx := r.GetCtx()
	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}
	convId := r.GetRouter("convId").String()

	g.Log().Debug(r.GetCtx(), "convId", convId)

	propsJson := gjson.New(Props)
	propsJson.Set("query.default.0", "c")

	propsJson.Set("query.default.1", convId)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)

	r.Response.WriteTpl(config.CacheBuildId+"/chat.html", g.Map{
		"props":       propsJson,
		"arkoseUrl":   config.ArkoseUrl,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(ctx),
	})
}

// Discovery 发现
func Discovery(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}

	propsJson := gjson.New(Props)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/gpts/discovery")

	r.Response.WriteTpl(config.CacheBuildId+"/discovery.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

// Gpts
func Gpts(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}

	propsJson := gjson.New(Props)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/gpts")

	r.Response.WriteTpl(config.CacheBuildId+"/gpts.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

// Editor 编辑器
func Editor(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}

	propsJson := gjson.New(Props)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/gpts/editor")

	// if slug != "" {
	// 	propsJson.Set("page", "/gpts/editor/[slug]")
	// 	propsJson.Set("query.slug", slug)
	// }
	// propsJson.Dump()

	r.Response.WriteTpl(config.CacheBuildId+"/editor.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

// Slug 编辑器
func Slug(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}
	slug := r.GetRouter("slug").String()

	propsJson := gjson.New(Props)

	propsJson.Set("query.slug", slug)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/gpts/editor/[slug]")

	r.Response.WriteTpl(config.CacheBuildId+"/slug.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

// G 游戏
func G(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}
	gizmoId := r.GetRouter("gizmoId").String()

	propsJson := gjson.New(PropsG)
	propsJson.Set("query.gizmoId", gizmoId)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/g/[gizmoId]")

	r.Response.WriteTpl(config.CacheBuildId+"/g.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

func GC(r *ghttp.Request) {

	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}
	gizmoId := r.GetRouter("gizmoId").String()
	convId := r.GetRouter("convId").String()
	g.Log().Debug(r.GetCtx(), "gizmoId", gizmoId)

	propsJson := gjson.New(Props)
	propsJson.Set("query.gizmoId", gizmoId)
	propsJson.Set("query.convId", convId)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("page", "/g/[gizmoId]/c/[convId]")

	r.Response.WriteTpl(config.CacheBuildId+"/gc.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})
}

// Mine 我的
func Mine(r *ghttp.Request) {
	if r.Session.MustGet("offical-session").IsEmpty() {
		r.Session.RemoveAll()
		r.Response.RedirectTo("/login")
		return
	}

	propsJson := gjson.New(Props)
	propsJson.Set("buildId", config.BuildId)
	propsJson.Set("assetPrefix", config.AssetPrefix)
	propsJson.Set("page", "/gpts/mine")
	r.Response.WriteTpl(config.CacheBuildId+"/mine.html", g.Map{
		"arkoseUrl":   config.ArkoseUrl,
		"props":       propsJson,
		"assetPrefix": config.AssetPrefix,
		"envScript":   config.GetEnvScript(r.GetCtx()),
	})

}
