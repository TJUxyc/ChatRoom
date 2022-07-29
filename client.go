package main

import (
	"fmt"
	"bufio"
	"net"
	"os"
	"strings"
)

var (
	connected bool 	= false
	done 			= make(chan struct{})
)

func main() {
	// 输入当前电脑的ip，如果只是想本地运行，不经过局域网可以直接输入127.0.0.1
	conn, err := net.Dial("tcp", "192.168.9.88:50000")
	if err != nil {
		fmt.Println("Cannot connect to the server", err.Error())
		return
	}
	defer conn.Close()

	inputReader := bufio.NewReader(os.Stdin)
	
	// 第一步，设置用户昵称
	fmt.Println("Please set your nickname: ")
	clientName, _ := inputReader.ReadString('\n')
	trimmedClient := strings.Trim(clientName, "\n")
	writeMessage(conn, "Login#" + trimmedClient)
	fmt.Println(recieveMessage(conn))
	// 预先获取一次当前的活跃用户列表
	writeMessage(conn, "List#")
	fmt.Println(recieveMessage(conn))

	// 进入聊天状态，开启接受对方发来的聊天信息的goroutine
	helpInfo := "==============================================================\n"
	helpInfo += "||Enter %\"{USER NAME}\" to sent your message.                ||\n"
	helpInfo += "||  |-> $\"{GROUP NAME}\" to send your message to the group.  ||\n"
	helpInfo += "||  |-> \"user list\" to flush the user list.                 ||\n"
	helpInfo += "||  |-> \"group list\" to get the group list.                 ||\n"
	helpInfo += "||  |-> \"create group\" to create the new group.             ||\n"
	helpInfo += "||  |-> \"join group\" to join the group.                     ||\n"
	helpInfo += "||  |-> \"exit group\" to exit the group.                     ||\n"
	helpInfo += "||  |-> \"quit\" to quit.                                     ||\n"
	helpInfo += "==============================================================\n"
	fmt.Println(helpInfo)
	connected = true
	go getMessage(conn)

	var (
		targetName	string
		msg			string
	)
	for {
		fmt.Printf("#To>")
		input, _ := inputReader.ReadString('\n')
		targetName = strings.Trim(input, "\n")
		// fmt.Println(targetName)
		if targetName == "quit" {
			writeMessage(conn, "Logout#")
			connected = false
			break
		} else if targetName == "user list" {
			writeMessage(conn, "List#")
		} else if targetName == "group list" {
			writeMessage(conn, "Group#")
		} else if targetName == "create group"{
			fmt.Printf("#Group Name>")
			input, _ = inputReader.ReadString('\n')
			msg = strings.Trim(input, "\n")
			writeMessage(conn, "Create#" + msg)
		} else if targetName == "join group"{
			fmt.Printf("#Group Name>")
			input, _ = inputReader.ReadString('\n')
			msg = strings.Trim(input, "\n")
			writeMessage(conn, "Join#" + msg)
		} else if targetName == "exit group"{
			fmt.Printf("#Group Name>")
			input, _ = inputReader.ReadString('\n')
			msg = strings.Trim(input, "\n")
			writeMessage(conn, "Exit#" + msg)
		} else if targetName[0] == '$'{
			// 发送到群组
			fmt.Printf("#Content>")
			input, _ = inputReader.ReadString('\n')
			msg = strings.Trim(input, "\n")
			writeMessage(conn, "Broadcast#" + targetName + "#" + msg)
		} else if targetName[0] == '%' {
			// 发送到某个个体
			targetName = targetName[1:]
			fmt.Printf("#Content>")
			input, _ = inputReader.ReadString('\n')
			msg = strings.Trim(input, "\n")
			writeMessage(conn, "Send#" + targetName + "#" + msg)
		} else {
			fmt.Println("Invaild Input!")
			continue
		}
		// 等待接受到消息后，再开启下轮循环
		<-done
	}
}

func getMessage(conn net.Conn) {
	for {
		if connected == false {
			break
		}
		data := recieveMessage(conn)
		if data != "success" {
			fmt.Println(data)
		}
		// 开头带#的是突发消息（即其他用户发给当前用户的消息），不需要进行同步。
		// 否则需要进行同步，以正常显示出系统命令行的内容
		if data[0] != '#' {
			done <- struct{}{}
		} else {
			fmt.Printf("#To>")
		}
	}
}

func recieveMessage(conn net.Conn) string{
	buf := make([]byte, 512)
	length, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading", err.Error())
		return "Error"
	}
	return string(buf[:length])
}

func writeMessage(conn net.Conn, sendMessage string) {
	if _, err := conn.Write([]byte(sendMessage)); err != nil {
		fmt.Println("Failed to send your message, please try again")
	}
	return
}