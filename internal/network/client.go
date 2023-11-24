package network

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type Client struct {
	conn   *websocket.Conn
	server *Server
	id     string
	egress chan []byte
}

func NewClient(conn *websocket.Conn, server *Server) *Client {
	id := uuid.NewString()
	fmt.Printf("Client connected (%s)\n", id)
	return &Client{
		conn:   conn,
		server: server,
		id:     id,
		egress: make(chan []byte),
	}
}

func (c *Client) ReadHandler() {
	defer func() {
		c.conn.Close()
		c.server.RemoveClient(c)
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			continue
		}

		if err := c.server.RouteEvent(&request, c); err != nil {
			log.Println("Error routing event Message: ", err)
		}
	}
}

func (c *Client) WriteHandler() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.server.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Connection closed: ", err)
				}
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) pongHandler(_ string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
