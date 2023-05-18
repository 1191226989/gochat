package service

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap     map[string]*User
	onlineMapLock sync.RWMutex

	//消息广播channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string, 0),
	}
	return server
}

func (s *Server) Handler(conn net.Conn) {
	defer conn.Close()
	// fmt.Println("连接建立成功")
	// 用户上线加到OnlineMap
	user := NewUser(conn, s)

	// 广播用户上线的消息
	user.Online()

	// 监听用户是否活跃的chan
	isLive := make(chan bool)
	defer close(isLive)

	cancelChan := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		for {
			select {
			case <-cancelChan:
				fmt.Println("select cancel")

				return
			default:
				// time.Sleep(time.Second * 1)
				buf := make([]byte, 4096)
				n, err := conn.Read(buf)
				if err != nil && err == io.EOF {
					return
				}
				if err != nil && err != io.EOF {
					fmt.Println("conn.Read err:", err)
					continue
				}
				if n == 0 {
					fmt.Println(n)
					// s.BroadCast(user, "下线")
					continue
				}

				// 提取用户的消息(去掉\n)
				msg := string(buf[:n-1])

				// 处理用户的消息
				user.HandleMessage(msg)

				// 用户的任意消息代表活跃
				isLive <- true
			}
		}
	}()

	// 判断是否活跃
	for {
		select {
		case <-isLive:

		case <-time.After(time.Second * 300):
			close(cancelChan)
			user.Offline()
			user.SendMessage("你已被迫离线\n")

			return
		}

	}

}

// 启动服务器
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close()

	// 启动监控Message的goroutine
	go s.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		go s.Handler(conn)
	}
}

// 监听message广播消息的channel
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		// 将收到的msg发给所有在线的user
		s.onlineMapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.Channel <- msg
		}
		s.onlineMapLock.Unlock()
	}
}

// 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s", user.Addr, user.Name, msg)
	s.Message <- sendMsg
}
