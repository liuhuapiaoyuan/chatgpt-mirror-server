package adminapi

import (
	"chatgpt-mirror-server/modules/chatgpt/service"
	"context"

	"github.com/cool-team-official/cool-admin-go/cool"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

type ChatgptUserController struct {
	*cool.Controller
}

func init() {
	var chatgpt_user_controller = &ChatgptUserController{
		&cool.Controller{
			Prefix:  "/adminapi/chatgpt/user",
			Api:     []string{"Add", "Delete", "Update", "Info", "List", "Page"},
			Service: service.NewChatgptUserService(),
		},
	}
	// 注册路由
	cool.RegisterController(chatgpt_user_controller)
}

// 增加 Welcome 演示 方法
type ChatgptUserWelcomeReq struct {
	g.Meta `path:"/welcome" method:"GET"`
}
type ChatgptUserWelcomeRes struct {
	*cool.BaseRes
	Data interface{} `json:"data"`
}

func (c *ChatgptUserController) Welcome(ctx context.Context, req *ChatgptUserWelcomeReq) (res *ChatgptUserWelcomeRes, err error) {
	res = &ChatgptUserWelcomeRes{
		BaseRes: cool.Ok("Welcome to Cool Admin Go"),
		Data:    gjson.New(`{"name": "Cool Admin Go", "age":0}`),
	}
	return
}

type AuthReq struct {
	g.Meta      `path:"/auth" method:"POST"`
	AccessToken string `json:"access_token"`
}
type AuthRes struct {
	*cool.BaseRes
	Data interface{} `json:"data"`
}

func (c *ChatgptUserController) Auth(ctx context.Context, req *AuthReq) (res *AuthRes, err error) {
	s := service.NewChatgptUserService()
	data, err := s.Auth(ctx, req.AccessToken)
	if err != nil {
		res = &AuthRes{
			BaseRes: cool.Fail(err.Error()),
		}
		return
	}
	res = &AuthRes{
		BaseRes: cool.Ok(""),
		Data:    data,
	}
	return
}
