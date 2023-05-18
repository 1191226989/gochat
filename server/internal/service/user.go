package service

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name    string
	Addr    string
	Channel chan string
	conn    net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		Channel: make(chan string),
		conn:    conn,
		server:  server,
	}

	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

func (u *User) ListenMessage() {
	for {
		msg := <-u.Channel
		u.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (u *User) Online() {
	u.server.onlineMapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.onlineMapLock.Unlock()

	// 广播用户上线消息
	u.server.BroadCast(u, "已上线\n")
}

// 用户离线
func (u *User) Offline() {
	u.server.onlineMapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.onlineMapLock.Unlock()

	// 广播用户离线消息
	u.server.BroadCast(u, "已离线\n")
}

// 处理用户消息
func (u *User) HandleMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户
		u.server.onlineMapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[%s]%s:在线...\n", user.Addr, user.Name)
			u.SendMessage(onlineMsg)
		}
		u.server.onlineMapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|张山
		newName := strings.Split(msg, "|")[1]
		// 判断name是否已经存在
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.SendMessage("当前用户名已经使用\n")
			return
		}

		u.server.onlineMapLock.Lock()
		delete(u.server.OnlineMap, u.Name)
		u.Name = newName
		u.server.OnlineMap[newName] = u
		u.server.onlineMapLock.Unlock()
		u.SendMessage("用户名更新成功\n")
	} else if msg[:3] == "to|" {
		// 消息格式 to|张山|hello world...
		name := strings.Split(msg, "|")[1]
		msg := strings.Split(msg, "|")[2]

		user, ok := u.server.OnlineMap[name]
		if !ok {
			u.SendMessage("用户名不存在\n")
			return
		}
		user.SendMessage(fmt.Sprintf("%s said:%s \n", u.Name, msg))

	} else {
		u.server.BroadCast(u, msg)
	}

}

// 用户发送消息
func (u *User) SendMessage(msg string) {
	u.conn.Write([]byte(msg))
}
