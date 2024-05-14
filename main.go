package main

import (
	_ "chatgpt-mirror-server/internal/packed"

	_ "github.com/cool-team-official/cool-admin-go/contrib/drivers/mysql"
	_ "github.com/cool-team-official/cool-admin-go/contrib/drivers/sqlite"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	_ "chatgpt-mirror-server/api"
	_ "chatgpt-mirror-server/arkose"
	_ "chatgpt-mirror-server/backend-api"
	_ "chatgpt-mirror-server/modules"
	_ "chatgpt-mirror-server/public-api"

	"github.com/gogf/gf/v2/os/gctx"

	"chatgpt-mirror-server/internal/cmd"
)

func main() {
	// gres.Dump()
	cmd.Main.Run(gctx.New())

	// 打印一段话： 服务已经启动，请访问 http://127.0.0.1
	//fmt.Println("服务已经启动，请访问 http://127.0.0.1")

}
