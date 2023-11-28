## 改造支持access_token的方案

1. 登录账号可以选择绑定的OPENAI账号
1.1 支持会话隔离，会话隔离需要增加表： 用户-会话表，自己创建的会话才会显示
: /backend-api/conversations
通过 backend-api/backend-api.go 判断

2. OPENAI账号支持使用 authorization的方式进行登陆，登录后需要设定过期时间，需要手工过期
3. 附件URL的统一拦截应该和会话隔离一样，处理backend-api.go/ProxyAll这个函数


# ChatGPT-Mirror-Server

## 快速部署脚本

[ChatGPT-Mirror-Server-Deploy](https://github.com/xyhelper/chatgpt-mirror-server-deploy)

## 集成跳转登陆接口

```
/login_token?access_token=9efe2cad-7534-45ef-b91d-xxxxxxxx
```