# ChatGPT-Mirror-Server

本服务为商业服务，自2024年4月16日0:00不再提供免费接入点。需付费使用。
1.流量大使用付费接入点或者小流量拼车付费接入点的，可打开以下链接选择适合自己的方案
https://xyhelper.cn/access

2.流量小或者小团队自用的，可使用蟑螂v1，目前蟑螂v1可正常使用，了解详情请访问:
CockroachAi（又名蟑螂）
https://github.com/cockroachai/cockroachai

## 快速部署脚本

[ChatGPT-Mirror-Server-Deploy](https://github.com/xyhelper/chatgpt-mirror-server-deploy)

## 集成跳转登陆接口

```
/login_token?access_token=9efe2cad-7534-45ef-b91d-xxxxxxxx
```


### v0.0.1
-  增加openai的账号模式， 账号密码自动登录模式，accessToken手工刷新模式
-  分配账号:增加账号密码
-  分配账号: 改为绑定openai账号（通过sessionId进行绑定）
-  允许登录的时候使用分配账号(注意不是openai账号)密码登录
-  [x] 增加会话隔离模式 isIsolate

### TODO 
- [] 增加GPTS的隔离
- [] 增加聊天对话过滤规则，每个账号可以配置一个过滤策略，过滤策略采用可视化编排，选择
