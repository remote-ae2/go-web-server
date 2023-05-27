package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MAX_CLIENTS = 100

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	readWait  = 10 * time.Minute

	pongWait   = 45 * time.Second
	pingPeriod = 40 * time.Second

	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = &Clients{w: make(map[string]*Client, 100), m: make(map[*Client]bool, 100)}

type Client struct {
	id string

	command string //последняя комманда

	conn *websocket.Conn

	rmchan bool //флаг для удаления канала
	send   chan interface{}
}

func (c *Client) remove() {
	if !c.rmchan {
		close(c.send)
		c.rmchan = true
	}
	clients.unregister(c)
	c.conn.Close()
}

func (c *Client) readPump() {
	defer func() {
		c.remove()
		log.Println("->readPump close")
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(readWait))
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Unmarshal error: %v", err)
			continue
		}

		//log.Printf("Message handle: [%s]\n", msg.Type)

		switch msg.Type {
		case "CompID":
			if msg.ID != "" {
				c.id = msg.ID
				//TODO проверить присылал уже OC клиент на id запросы
				clients.wait(c, c.id)
				c.send <- createWaitingMessage()
			}

		default:
			log.Printf("Message type error: [%s]\n", msg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.remove()
		log.Println("<-writePump close")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				//канал закрыт
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			switch msg := message.(type) {
			case string:
				w.Write([]byte(msg))

			default:
				//json
				b, err := json.Marshal(message)
				if err != nil {
					log.Println(err.Error())
					return
				}
				w.Write(b)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleHttp(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		//log.Println("GET request received")

		keys := r.URL.Query()
		id := keys.Get("id")
		session := keys.Get("session")

		//TODO проверка id, session

		//ждущий клиент
		client := clients.getClient(id)
		if client == nil {
			fmt.Fprintf(w, "end")
			return
		}

		if session == "false" {
			log.Println("Client found; start session")
			fmt.Fprintf(w, "session")
			//сообщить клиенту о начале сессии
			client.send <- createSessionStartMessage()
		}
		fmt.Fprintf(w, client.command)

	case "POST":
		//log.Println("POST request received")

		if err := r.ParseForm(); err != nil {
			log.Printf("ParseForm() err: %v", err)
			return
		}

		var storage StorageAE2
		var data string
		for data, _ = range r.Form {
			//log.Println(data)

			err := json.Unmarshal([]byte(data), &storage)
			if err != nil {
				log.Println(err.Error())
				return
			}
		}
		//log.Println(storage.Items[1].Name)
		//log.Println(data)

		//TODO проверка ID

		//ждущий клиент
		client := clients.getClient(storage.ID)
		if client != nil {
			client.send <- data
		}

	default:
	}
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if clients.size() >= MAX_CLIENTS {
		log.Println("Cannot handle more requests")
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan interface{}, 256),
	}
	clients.register(client)

	go client.writePump()
	go client.readPump()

	log.Println("Websocket connection received")
}
