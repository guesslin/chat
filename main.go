package main

import (
	"bufio"
	"fmt"
	"net"
)

var NumofClients int = 1

type Client struct {
	id        string
	incomming chan string
	outgoing  chan string
	reader    *bufio.Reader
	writer    *bufio.Writer
}

func (c *Client) Read() {
	for {
		line, _ := c.reader.ReadString('\n')
		if line != "" {
			c.incomming <- fmt.Sprintf("%s: %s", c.id, line)
		}
	}
}

func (c *Client) Write() {
	for data := range c.outgoing {
		c.writer.WriteString(data)
		c.writer.Flush()
	}
}

func (c *Client) Listen() {
	go c.Read()
	go c.Write()
}

func (c *Client) Close() {
}

func NewClientID() string {
	name := fmt.Sprintf("%d", NumofClients)
	NumofClients++
	return name
}

func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	client := &Client{
		id:        NewClientID(),
		incomming: make(chan string),
		outgoing:  make(chan string),
		reader:    reader,
		writer:    writer,
	}
	fmt.Println(client.id, "is joins")
	client.Listen()
	return client
}

type ChatRoom struct {
	clients   []*Client
	join      chan net.Conn
	incomming chan string
	outgoing  chan string
}

func (c *ChatRoom) Broadcast(data string) {
	for _, client := range c.clients {
		client.outgoing <- data
	}
}

func (c *ChatRoom) Join(conn net.Conn) {
	client := NewClient(conn)
	c.clients = append(c.clients, client)
	go func() {
		for {
			c.incomming <- <-client.incomming
		}
	}()
}

func (c *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case data := <-c.incomming:
				c.Broadcast(data)
			case conn := <-c.join:
				c.Join(conn)
			}
		}
	}()
}

func (c *ChatRoom) CleanClients() {
	go func() {
		for client := range c.clients {
			_ = client
		}
	}()
}

func NewChatRoom() *ChatRoom {
	chatroom := &ChatRoom{
		clients:   make([]*Client, 0),
		join:      make(chan net.Conn),
		incomming: make(chan string),
		outgoing:  make(chan string),
	}
	chatroom.Listen()
	return chatroom
}

func main() {
	chatroom := NewChatRoom()
	listener, _ := net.Listen("tcp", ":5555")
	fmt.Println("Server host on localhost:5555")
	for {
		conn, _ := listener.Accept()
		chatroom.join <- conn
	}
}
