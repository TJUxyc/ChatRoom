## TCP局域网点对点聊天/群聊小程序
> version: 3.0 
> by Xu Yicheng

#### 使用方法
* go run server.go // 运行服务器程序
* go run client.go // 运行客户端程序
#### 核心知识点：Golang下的tcp通信（net库的使用）
* net.Listen()
* net.Dial()
* listner.Accept()
* conn.Write()
* conn.Read()
#### TCP层通信协议设计
1. **请求报文**
格式: {**COMMAND**}**#**{**DATA**}
* COMMAND: 通信指令
* DATA: 消息正文

| COMMAND   | DATA                       | Describe         |
|-----------|----------------------------|------------------|
| Login     | User name                  | 登陆系统|
| List      | nil                        | 获取用户列表|  
| Group     | nil                        | 获取群组列表|
| Create    | Group name                 | 创建群组|
| Join      | Group name                 | 加入群组|
| Exit      | Group name                 | 退出群组|
| Send      | User name#Message content  | 发送私聊消息|
| Broadcast | Group name#Message content | 发送群聊消息|
| Logout    | nil                        | 登出系统|
| else      | (Invalid message)          | 非法指令|
2. **响应报文**
格式：**#**(optional)+**DATA**
* 最前面带#的是突发消息，指其他用户直接或者间接（通过group）发送给用户的消息。
* 没带#的是服务器回应给客户端的响应消息，服务端每次收到一条客户端发来的消息，都会做出一条响应。
#### 用户界面设计
1. 点对点聊天用例
输入用户名--->连接到聊天系统--->查看当前用户列表--->输入用户名称--->输入聊天信息--->发送私聊信息--->退出系统
2. 群聊用例
输入用户名--->连接到聊天系统--->查看当前群组列表--->创建某个群组（加入某个群组）--->输入群组名称--->输入群聊信息--->发送群聊信息--->退出群组--->退出系统
3. 用户指令设计

| #To(Fist line) | Second line name  | Second line content | Desribe |
|----------------|-------------------|---------------------|---------|
| %{User name}   | #Content          | Message content     | 私聊 |
| ${Group name}  | #Content          | Message content     | 群聊 |
| user list      | /                 |                     | 获取用户列表|
| group list     | /                 |                     | 获取群组列表|
| create group   | #Group name       | group name          | 创建群组 |
| join group     | #Group name       | group name          | 加入群组 |
| exit group     | #Group name       | group name          | 退出群组 |
| quit           | /                 |                     | 退出系统 |
| else           | (Invalid command) |                     | 非法指令 |
#### 待完善之处
* 服务器ip地址硬编码到了代码中，可以把ip放到可执行文件的命令行参数里面。
* 系统中以user name为唯一标识，但是人名可能出现重复的情况，此时前面的用户会被覆盖。
* 当前的用户指令系统操作起来比较复杂，例如每次私聊都要先完整地输入对方的用户名；可以通过%和$后输入空串，默认发给上一个用户/群组的方式缩短通信的流程。
* 图像化界面，支持文件的收发，高并发（看自己想实现的功能了）
