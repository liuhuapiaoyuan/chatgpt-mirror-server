package config

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gview"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

//	func CHATPROXY(ctx g.Ctx) string {
//		return g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
//	}
func WS_SERVICE(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "WS_SERVICE").String()
}

func AUTHKEY(ctx g.Ctx) string {
	// g.Log().Debug(ctx, "config.AUTHKEY", g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String())
	return g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String()
}

func USERTOKENLOCK(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "USERTOKENLOCK").Bool()
}

var (
	CHATPROXY    = "https://chatproxy-dev.closeai.biz"
	DefaultModel = "text-davinci-002-render-sha"
	FreeModels   = garray.NewStrArray()
	Ja3Proxy     *url.URL // ja3代理
	PlusModels   = garray.NewStrArray()
	ArkoseUrl    = "/v2/"
	BuildId      = "MCkVH1jJi3yNLkMToVDdU"
	CacheBuildId = "MCkVH1jJi3yNLkMToVDdU"
	AssetPrefix  = "https://oaistatic-cdn.closeai.biz"
	PK40         = "35536E1E-65B4-4D96-9D97-6ADB7EFF8147"
	ProxyClient  *gclient.Client
	PK35         = "3D86FBBA-9D22-402A-B512-3420086BA6CC"
	envScriptTpl = `
	<script>
	window.__arkoseUrl="{{.ArkoseUrl}}";
	window.__assetPrefix="{{.AssetPrefix}}";
	window.__PK40="{{.PK40}}";
	window.__PK35="{{.PK35}}";
	</script>
	`
)

func init() {
	ctx := gctx.GetInitCtx()
	ProxyClient = g.Client().SetBrowserMode(true).SetHeaderMap(g.MapStrStr{
		"Origin":     "https://chat.openai.com",
		"Referer":    "https://chat.openai.com/",
		"Host":       "chat.openai.com",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	})
	// 读取ja3代理
	ja3Proxy := g.Cfg().MustGetWithEnv(ctx, "JA3_PROXY").String()
	if ja3Proxy != "" {

		u, err := url.Parse(ja3Proxy)
		if err != nil {
			panic(err)
		}
		Ja3Proxy = u
		g.Log().Info(ctx, "JA3_PROXY:", Ja3Proxy.String())
		ProxyClient = g.Client().Proxy(Ja3Proxy.String()).SetBrowserMode(true).SetHeaderMap(g.MapStrStr{
			"Origin":     "https://chat.openai.com",
			"Referer":    "https://chat.openai.com/",
			"Host":       "chat.openai.com",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		})
	}

	FreeModels.Append("text-davinci-002-render-sha")
	FreeModels.Append("text-davinci-002-render-sha-mobile")
	PlusModels.Append("gpt-4")
	PlusModels.Append("gpt-4-browsing")
	PlusModels.Append("gpt-4-plugins")
	PlusModels.Append("gpt-4-mobile")
	PlusModels.Append("gpt-4-gizmo")
	// arkoseUrl := g.Cfg().MustGetWithEnv(ctx, "ARKOSE_URL")
	// if !arkoseUrl.IsEmpty() {
	// 	ArkoseUrl = arkoseUrl.String()
	// }
	chatproxy := g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
	if chatproxy != "" {
		CHATPROXY = chatproxy
	}
	g.Log().Info(ctx, "CHATPROXY:", CHATPROXY)
	assetPrefix := g.Cfg().MustGetWithEnv(ctx, "ASSET_PREFIX").String()
	if assetPrefix != "" {
		AssetPrefix = assetPrefix
	}
	g.Log().Info(ctx, "ASSET_PREFIX:", AssetPrefix)
	cacheBuildId := CheckVersion(ctx, AssetPrefix)
	if cacheBuildId != "" {
		CacheBuildId = cacheBuildId
	}
	g.Log().Info(ctx, "CacheBuildId:", CacheBuildId)
	build := CheckNewVersion(ctx)
	if build != "" {
		BuildId = build
	}
	g.Log().Info(ctx, "BuildId:", BuildId)
	// 每小时更新一次
	go func() {
		for {
			time.Sleep(time.Hour)
			build := CheckNewVersion(ctx)
			if build != "" {
				BuildId = build
			}
			g.Log().Info(ctx, "BuildId:", BuildId)
			cacheBuildId := CheckVersion(ctx, AssetPrefix)
			if cacheBuildId != "" {
				CacheBuildId = cacheBuildId
			}
			g.Log().Info(ctx, "CacheBuildId:", CacheBuildId)
		}
	}()

}

func PORT(ctx g.Ctx) int {
	g.Log().Debug(ctx, "config.PORT", g.Cfg().MustGetWithEnv(ctx, "PORT").Int())
	if g.Cfg().MustGetWithEnv(ctx, "PORT").Int() == 0 {
		return 8001
	}
	return g.Cfg().MustGetWithEnv(ctx, "PORT").Int()
}

func ONLYTOKEN(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "ONLYTOKEN").Bool()
}

// 获取cdnhost
func CDNHOST(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "CDNHOST").String()
}

func CRONINTERVAL(ctx g.Ctx) string {
	// 生成随机时间的每3天执行一次的表达式，格式为：秒 分 时 天 月 星期
	// 生成随机秒数 在0-59之间
	second := generateRandomNumber(59)
	secondStr := gconv.String(second)
	// 生成随机分钟数 在0-59之间
	minute := generateRandomNumber(59)
	minuteStr := gconv.String(minute)
	// 生成随机小时数 在0-23之间
	hour := generateRandomNumber(23)
	hourStr := gconv.String(hour)
	// 拼接cron表达式
	cronStr := secondStr + " " + minuteStr + " " + hourStr + " * * *"
	return cronStr

}

func generateRandomNumber(max int) int {
	rand.Seed(time.Now().UnixNano()) // 使用当前时间作为随机数生成器的种子
	return rand.Intn(max)            // 生成0到59之间的随机数
}

func APIAUTH(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "APIAUTH").String()
}

// 是否在新绑定用户时清空聊天记录
func CLEARCHATHISTORY(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "CLEARCHATHISTORY").Bool()
}

// 检查版本号并同步资源
func CheckVersion(ctx g.Ctx, assetPrefix string) (CacheBuildId string) {
	gclient := g.Client()
	// 读取 assetPrefix/version
	versionVar := gclient.GetVar(ctx, assetPrefix+"/version.json")
	CacheBuildId = gjson.New(versionVar).Get("cacheBuildId").String()
	g.Log().Infof(ctx, "Get config From %s ,CacheBuildId: %s", AssetPrefix, CacheBuildId)
	if CacheBuildId == "" {
		return ""
	}
	// 读取buildDate目录索引
	indexUrl := assetPrefix + "/template/" + CacheBuildId + "/index.txt"
	g.Log().Info(ctx, "Get config From ", indexUrl)
	buildDateVar := gclient.GetVar(ctx, indexUrl).String()
	if buildDateVar == "" {
		return ""
	}
	// 按回车分割
	buildDateList := gstr.Split(buildDateVar, "\n")
	g.Dump(buildDateList)
	// 遍历目录索引 如果没有就下载
	for _, v := range buildDateList {
		if v == "" {
			continue
		}
		// 检查文件是否存在
		if !gfile.Exists("./resource/template/" + CacheBuildId + "/" + v) {
			g.Log().Infof(ctx, "Download %s", v)
			// 下载文件
			res, err := gclient.Get(ctx, assetPrefix+"/template/"+CacheBuildId+"/"+v)
			if err != nil {
				g.Log().Error(ctx, "Download  Error: ", v, err)
				return ""
			}
			defer res.Close()
			if res.StatusCode != 200 {
				g.Log().Error(ctx, "Download  Error: ", v, res.StatusCode)
				return ""
			}
			// 写入文件
			err = gfile.PutBytes("./resource/template/"+CacheBuildId+"/"+v, res.ReadAll())
			if err != nil {
				g.Log().Error(ctx, "Download  Error: ", v, err)
				return ""
			}

		}
	}

	return
}

func GetEnvScript(ctx g.Ctx) string {
	script, err := gview.ParseContent(ctx, envScriptTpl, g.Map{
		"ArkoseUrl":   ArkoseUrl,
		"AssetPrefix": AssetPrefix,
		"PK40":        PK40,
		"PK35":        PK35,
	})
	if err != nil {
		g.Log().Error(ctx, "GetEnvScript Error: ", err)
		return ""
	}
	return script
}

// 检查是否有新版本
func CheckNewVersion(ctx g.Ctx) (buildId string) {
	resVar := g.Client().GetVar(ctx, CHATPROXY+"/ping")
	resJson := gjson.New(resVar)

	buildId = resJson.Get("buildId").String()
	return
}
