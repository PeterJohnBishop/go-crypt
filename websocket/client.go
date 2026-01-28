package websocket

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xlzd/gotp"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	id     string
	alias  string
	secret string
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	allow  []string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
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

// generate a uri for totp
func generateTOTPWithSecret(clientId string, clientSecret string, alias string) string {
	totp := gotp.NewDefaultTOTP(clientSecret)

	issuer := fmt.Sprintf("direct_message:%s", alias)

	uri := totp.ProvisioningUri(clientId, issuer)

	return uri
}

// verify incomming OTP
func verifyOTP(randomSecret string, otp string) {
	totp := gotp.NewDefaultTOTP(randomSecret)

	if totp.Verify(otp, time.Now().Unix()) {
		fmt.Println("Authentication successful! Access granted.")
	} else {
		fmt.Println("Authentication failed! Invalid OTP.")
	}
}

func GetSecretFromDB(id string) string {
	return "under construction"
}

func ServeWs(hub *Hub, c *gin.Context) {

	alias := c.Param("alias")
	if alias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alias is required"})
		return
	}

	id := c.GetHeader("X-Client-Id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Client-Id header is missing"})
		return
	}

	var secret string
	var isNew bool

	secret = GetSecretFromDB(id)
	if secret != "" {
		isNew = false
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if secret == "" {
		secretLength := 16
		secret = gotp.RandomSecret(secretLength)
		isNew = true
	}

	if isNew {
		data := generateTOTPWithSecret(id, secret, alias)
		msg := map[string]string{
			"type": "totp dm verification",
			"data": data,
		}
		conn.WriteJSON(msg)
	}

	client := &Client{id: id, alias: alias, secret: secret, hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
