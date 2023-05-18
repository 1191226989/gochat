package service

import (
	"fmt"
	"io"
	"net"
	"os"
)

type client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前客户端的模式
}

func NewClient(serverIp string, serverPort int) (*client, error) {
	// 创建客户端对象
	client := &client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil, err
	}

	client.conn = conn
	return client, err
}

func (c *client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("请输入0-3的数字")
		return false
	}
}

func (c *client) Run() {
	for {
		b := c.menu()
		if b == false {
			continue
		}

		switch c.flag {
		case 0:
			fmt.Println("退出")
			return
		case 1:
			fmt.Println("公聊模式")
			c.PublicChat()
		case 2:
			fmt.Println("私聊模式")
			c.PrivateChat()
		case 3:
			fmt.Println("更新用户名")
			ret := c.UpdateName()
			if ret == false {
				fmt.Println("更新用户名失败")
			} else {
				fmt.Println("更新用户名成功")
			}
		}
	}

}

func (c *client) UpdateName() bool {
	fmt.Println("请输入新的用户名:")
	fmt.Scanln(&c.Name)

	sendMsg := fmt.Sprintf("rename|%s\n", c.Name)
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (c *client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
}

func (c *client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for {
		if chatMsg == "exit" {
			return
		}
		if len(chatMsg) != 0 {
			_, err := c.conn.Write([]byte(chatMsg + "\n"))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (c *client) ShowUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (c *client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.ShowUsers()
	fmt.Println(">>>>请输入聊天对象的[用户名],exit退出.")
	fmt.Scanln(&remoteName)

	for {
		if remoteName == "exit" {
			fmt.Println("退出")
			return
		}

		fmt.Println(">>>>请输入消息内容,exit退出.")
		fmt.Scanln(&chatMsg)

		for {
			if chatMsg == "exit" {
				fmt.Println("退出")
				break
			}

			if len(chatMsg) != 0 {
				sendMsg := fmt.Sprintf("to|%s|%s\n", remoteName, chatMsg)

				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}

			}

			chatMsg = ""
			fmt.Println("请输入消息内容,exit退出.")
			fmt.Scanln(&chatMsg)
		}

		c.ShowUsers()
		fmt.Println(">>>>请输入聊天对象的[用户名],exit退出.")
		fmt.Scanln(&remoteName)
	}
}
