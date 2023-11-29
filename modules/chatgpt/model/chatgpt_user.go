package model

import (
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
)

const TableNameChatgptUser = "chatgpt_user"

// ChatgptUser mapped from table <chatgpt_user>
type ChatgptUser struct {
	*cool.Model
	UserToken  string     `gorm:"column:userToken;not null;comment:UserToken" json:"userToken"`
	username   string     `gorm:"column:username;not null;comment:登录账号" json:"username"`
	password   string     `gorm:"column:password;not null;comment:登录密码" json:"password"`
	ExpireTime *time.Time `gorm:"column:expireTime;not null;comment:过期时间" json:"expireTime"`
	IsPlus     bool       `gorm:"column:isPlus;comment:PLUS;default:0" json:"isPlus"`
	isIsolate  bool       `gorm:"column:isIsolate;comment:会话隔离;default:0" json:"isIsolate"`
	Remark     string     `gorm:"column:remark;comment:备注" json:"remark"`
	SessionId  int        `gorm:"column:remark;comment:绑定账号" json:"sessionId"`
}

// TableName ChatgptUser's table name
func (*ChatgptUser) TableName() string {
	return TableNameChatgptUser
}

// GroupName ChatgptUser's table group
func (*ChatgptUser) GroupName() string {
	return "default"
}

// NewChatgptUser create a new ChatgptUser
func NewChatgptUser() *ChatgptUser {
	return &ChatgptUser{
		Model: cool.NewModel(),
	}
}

// init 创建表
func init() {
	cool.CreateTable(&ChatgptUser{})
}
