package main

import (
	"fmt"
	"net"
	"strings"
)

type Mes struct {
	targetIp, content string
}

type void struct{}
type set map[string]void

var (
	// 用户ip地址 ----> 用户昵称
	ip2Name 	= make(map[string]string)
	// 用户昵称   ----> tcp连接
	name2Conn	= make(map[string]net.Conn)
	// 群组名称   ----> 用户列表（存放用户的昵称）
	group		= make(map[string]set)
)

func main() {
	fmt.Println("Starting the server ...")
	// 创建监听器
	listener, err := net.Listen("tcp", "192.168.9.88:50000")
	if err != nil {
		fmt.Println("Error listening", err.Error())
		return
	}
	// 监听并接受来自客户端的连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			return
		}
		go doServerStuff(conn)
	}
}

func doServerStuff(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	ip2Name[clientAddr] = "unset"
	usrName := "unset"

	// 如果服务器中没有对应的客户机ip地址，
	defer conn.Close()
	
	for {
		buf := make([]byte, 512)
		length, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading", err.Error())
			return
		}
		
		var (
			msg 		= string(buf[:length])			// 用户发来的信息
			command		= strings.Split(msg, "#")[0]	// “#”前为指令名称
			data		= strings.Split(msg, "#")[1]	// “#”后为具体的数据
		)
		switch command {
		// 用户设置昵称，设置完成后即可连接到聊天系统
		case "Login":
			usrName = data
			sendMessage := fmt.Sprintf("(System message)Hello, %s!\nYour ip address is: %s", usrName, clientAddr)
			ip2Name[clientAddr] = usrName
			name2Conn[usrName]  = conn
			writeMes(conn, sendMessage)
			defer closeConn(conn, clientAddr, usrName)
		// 获取当前在线用户列表
		case "List":
			sendMessage := fmt.Sprintf("============================================\n(System message)Total users: %d people", len(ip2Name))
			sendMessage += ("\nIp address\t\tUser name\n")
			for k, v := range ip2Name {
				if k == clientAddr {
					sendMessage += (k + "\t" + v + "(You)\n")
				} else {
					sendMessage += (k + "\t" + v + "\n")
				}
			}
			sendMessage += ("============================================")
			writeMes(conn, sendMessage)
		// 获取当前系统下的群组列表
		case "Group":
			sendMessage := fmt.Sprintf("============================================\n(System message)Total groups: %d group", len(group))
			sendMessage += ("\nGroup name\t\tMembers' name\n")
			for k, v := range group {
				if inGroup(v, usrName) {
					sendMessage += (k + "(followed)\t" + getSetElement(v) + "\n")
				} else {
					sendMessage += (k + "\t\t\t" + getSetElement(v) + "\n")
				}
			}
			sendMessage += ("============================================")
			writeMes(conn, sendMessage)
		// 创建群组
		case "Create":
			var (
				groupName	= data
				usrLs		= set{usrName: {}}
			)
			// 群组名是否已经被使用
			if _, ok := group[groupName]; ok {
				sendMessage := fmt.Sprintf("(System message)Group name %s has been used! Please try again", groupName)
				writeMes(conn, sendMessage)
			} else {
				group[groupName] = usrLs
				sendMessage := fmt.Sprintf("(System message)Create group %s successfully!", groupName)
				writeMes(conn, sendMessage)
			}
		// 加入群组
		case "Join":
			var (
				groupName	= data
				usrLs		= set{usrName: {}}
			)
			// 群组是否存在
			if _, ok := group[groupName]; !ok {
				sendMessage := fmt.Sprintf("(System message)Group name %s does not exist! Please check your input.", groupName)
				writeMes(conn, sendMessage)
			} else {
				usrLs = group[groupName]
				// 为群组添加新用户
				usrLs[usrName] = void{}
				group[groupName] = usrLs
				sendMessage := fmt.Sprintf("(System message)Join group %s successfully!", groupName)
				writeMes(conn, sendMessage)
			}
		// 退出群组
		case "Exit":
			var (
				groupName	= data
				usrLs		= set{}
			)
			// 群组是否存在
			if _, ok := group[groupName]; !ok {
				sendMessage := fmt.Sprintf("(System message)Group name %s does not exist! Please check your input.", groupName)
				writeMes(conn, sendMessage)
			} else {
				usrLs = group[groupName]
				// 用户是否在该群组当中
				if inGroup(usrLs, usrName) {
					// 是不是当前群组的最后一个人
					if len(usrLs) == 1 {
						// 直接删除该群组
						delete(group, groupName)
					} else {
						delete(usrLs, usrName)
					}
					sendMessage := fmt.Sprintf("(System message)Exit group %s successfully!", groupName)
					writeMes(conn, sendMessage)
				} else {
					sendMessage := fmt.Sprintf("(System message)Sorry, you are not in group %s.", groupName)
					writeMes(conn, sendMessage)
				}
			}
		// 发送信息（点对点）
		case "Send":
			var (
				targetName 	= data
				cont		= strings.Split(msg, "#")[2]
				sendMessage	= fmt.Sprintf("\r(%s)%s", usrName, cont)
			)
			// 判断系统中是否存在该用户
			if sendConn, ok := name2Conn[targetName]; ok {
				// 是否发给自己
				if targetName == usrName {
					sendMessage = fmt.Sprintf("(System message)Can't sent message to yourself!")
					writeMes(conn, sendMessage)
				} else {
					if _, err := sendConn.Write([]byte("#" + sendMessage)); err != nil {
						writeMes(conn, "Send failed")
					} else {
						writeMes(conn, "success")
					}
				}
			} else {
				sendMessage = fmt.Sprintf("(System message)User %s does not exist!", targetName)
				writeMes(conn, sendMessage)
			}
		// 发送消息（群聊）
		case "Broadcast":
			var (
				targetGroup	= data[1:]
				cont		= strings.Split(msg, "#")[2]
				sendMessage	= fmt.Sprintf("\r[%s](%s)%s", targetGroup, usrName, cont)
			)
			// 是否存在该群组
			if _, ok := group[targetGroup]; !ok {
				sendMessage = fmt.Sprintf("(System message)Group name %s does not exist! Please check your input.", targetGroup)
				writeMes(conn, sendMessage)
			} else {
				// 用户是否在该群组中
				if !inGroup(group[targetGroup], usrName) {
					sendMessage = fmt.Sprintf("(System message)Sorry, you are not in group %s.", targetGroup)
					writeMes(conn, sendMessage)
				} else {
					usrLs := group[targetGroup]
					// 群组中只有一个人（自己）
					if len(usrLs) <= 1 {
						sendMessage = fmt.Sprintf("(System message)Group %s only has you now, find more friend to join us.^_^", targetGroup)
						writeMes(conn, sendMessage)
						break
					}
					for k, _ := range usrLs {
						// 发送给除了自己的全部成员
						if k != usrName {
							sendConn := name2Conn[k]
							writeMes(sendConn, "#" + sendMessage)
						}
					}
					writeMes(conn, "success")
				}
			}
		// 断开和服务器的连接
		case "Logout":
			closeConn(conn, clientAddr, usrName)
		default:
			writeMes(conn, "Invaild format")
		}
	}
}

func writeMes(conn net.Conn, sendMessage string) {
	if _, err := conn.Write([]byte(sendMessage)); err != nil {
		fmt.Printf("Failed to send message (%s), please try again", sendMessage)
	}
	return
}

func closeConn(conn net.Conn, addr string, name string) {
	delete(ip2Name, addr)
	delete(name2Conn, name)
	// conn.Close()
}

func inGroup(members set, name string) bool {
	if _, ok := members[name]; ok {
		return true
	}
	return false
}

func getSetElement(member set) string {
	res := []string{}
	for k, _ := range member {
		res = append(res, k)
	}
	return strings.Join(res, ",")
}