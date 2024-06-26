package model

import (
	"github.com/cool-team-official/cool-admin-go/cool"
)

const TableNameChatgptSession = "chatgpt_session"

// ChatgptSession mapped from table <chatgpt_session>
type ChatgptSession struct {
	*cool.Model
	Email           string `gorm:"column:email;not null;comment:邮箱" json:"email"`
	Password        string `gorm:"column:password;not null;comment:密码" json:"password"`
	Status          bool   `gorm:"column:status;comment:状态;default:0" json:"status"`
	IsPlus          bool   `gorm:"column:isPlus;comment:PLUS;default:0" json:"isPlus"`
	UserID          int    `gorm:"column:userID;comment:用户ID" json:"userID"`
	Mode            int    `gorm:"column:mode;comment:模式(0=账号,1=token)" json:"mode"`
	OfficialSession string `gorm:"column:officialSession;comment:官方session" json:"officialSession"`
	Remark          string `gorm:"column:remark;comment:备注" json:"remark"`
}

// TableName ChatgptSession's table name
func (*ChatgptSession) TableName() string {
	return TableNameChatgptSession
}

// GroupName ChatgptSession's table group
func (*ChatgptSession) GroupName() string {
	return "default"
}

// NewChatgptSession create a new ChatgptSession
func NewChatgptSession() *ChatgptSession {
	return &ChatgptSession{
		Model: cool.NewModel(),
	}
}

// init 创建表
func init() {
	cool.CreateTable(&ChatgptSession{})
}
