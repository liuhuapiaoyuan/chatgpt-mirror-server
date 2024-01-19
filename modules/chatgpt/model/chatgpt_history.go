package model

import (
	"github.com/cool-team-official/cool-admin-go/cool"
)

const TableNameChatgptHistory = "chatgpt_history"

// ChatgptHistory mapped from table <chatgpt_history>
type ChatgptHistory struct {
	*cool.Model
	UserId                  int    `gorm:"column:user_id;not null;comment:用户ID" json:"userId"`
	Title                   string `gorm:"column:title;not null;comment:标题" json:"title"`
	ConversationId          string `gorm:"column:conversation_id;not null;unique;comment:会话ID" json:"conversationId"`
	ChatgptId               int    `gorm:"column:chatgpt_id;comment:CHATGPT账号ID" json:"chatgptId"`
	ConversationTemplate_id string `gorm:"column:conversation_template_id;comment:模板ID" json:"conversationTemplateId"`
	GizmoId                 string `gorm:"column:gizmo_id;comment:工具ID" json:"gizmoId"`
}

// TableName ChatgptHistory's table name
func (*ChatgptHistory) TableName() string {
	return TableNameChatgptHistory
}

// GroupName ChatgptHistory's table group
func (*ChatgptHistory) GroupName() string {
	return "default"
}

// NewChatgptHistory create a new ChatgptHistory
func NewChatgptHistory() *ChatgptHistory {
	return &ChatgptHistory{
		Model: cool.NewModel(),
	}
}

// init 创建表
func init() {
	cool.CreateTable(&ChatgptHistory{})
}
