package network

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Server struct {
	clients map[string]*Client
	sync.RWMutex
	listeners map[string]EventListener
}

func NewServer() *Server {
	s := &Server{
		clients:   map[string]*Client{},
		RWMutex:   sync.RWMutex{},
		listeners: map[string]EventListener{},
	}
	s.wireListeners()
	return s
}

func (s *Server) wireListeners() {
	s.listeners[EventMessage] = func(event *Event, c *Client) error {
		var message MessageEvent
		err := json.Unmarshal(event.Payload, &message)
		if err != nil {
			return err
		}

		if client, ok := s.clients[message.To]; ok {
			buff, err := json.Marshal(event)
			if err != nil {
				return err
			}
			client.egress <- buff
		}

		return nil
	}
}

func (s *Server) RouteEvent(event *Event, c *Client) error {
	if listener, ok := s.listeners[event.Type]; ok {
		if err := listener(event, c); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) AddClient(c *Client) {
	s.Lock()
	defer s.Unlock()
	s.clients[c.id] = c
}

func (s *Server) RemoveClient(c *Client) {
	s.Lock()
	defer s.Unlock()
	delete(s.clients, c.id)
}

func (s *Server) ListenAndServe(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := NewClient(conn, s)
		s.AddClient(c)

		go c.ReadHandler()
		go c.WriteHandler()
	})

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
