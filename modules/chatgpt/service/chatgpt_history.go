package service

import (
	"chatgpt-mirror-server/modules/chatgpt/model"

	"github.com/cool-team-official/cool-admin-go/cool"
)

type ChatgptHistoryService struct {
	*cool.Service
}

func NewChatgptHistoryService() *ChatgptHistoryService {
	return &ChatgptHistoryService{
		&cool.Service{
			Model: model.NewChatgptHistory(),
		},
	}
}

// 接受参数 userId和conversationIds，返回对应的历史记录的列表
func (s *ChatgptHistoryService) GetHistoryByUserIdAndConversationIds(userId string, conversationIds []string) (data interface{}, err error) {
	// 查询历史记录
	records, err := cool.DBM(s.Model).Where("user_id=?", userId).Where("conversation_id in (?)", conversationIds).All()
	if err != nil {
		return
	}
	return records, nil
}
