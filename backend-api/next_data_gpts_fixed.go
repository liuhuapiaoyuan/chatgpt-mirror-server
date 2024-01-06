package backendapi

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
)

func NextDataGptsFixed(r *ghttp.Request) {

	resJson := gjson.New(`
	{  }
`)
	r.Response.WriteJson(resJson)
}
