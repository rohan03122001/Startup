package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rohan03122001/quizzing/internal/websockets"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
		//need to work for production
	},
}


type WebSocketHandler struct{
	hub *websockets.Hub
}

func NewWebSocketHandler(hub *websockets.Hub) *WebSocketHandler{
	return &WebSocketHandler{
		hub: hub,
	}
}


//Handle HTTP to upgrade

func (h *WebSocketHandler)HandleHTTP(c *gin.Context){	
	roomID := c.Query("room")
	if roomID == ""{
		 c.JSON(http.StatusBadRequest, gin.H{
			"error": "room ID Required",
		})
		return
	}


	//Upgrade Connection to websocker

	conn, err := upgrader.Upgrade(c.Writer, c.Request,nil)
	if err!= nil{
		log.Printf("Failed to upgrade Connection %v", err)
		return
	}

	//Generate CLIENT ID FOR THE CONNECTION WHICH UPGRADED

	clientID := uuid.New().String()

	//create new client
	client := websockets.NewClient(h.hub, conn, roomID, clientID)

	// register with hub
	h.hub.Register <- client	

	//start go routines for read and write pump

	go client.ReadPump()
	go client.WritePump()

	log.Printf("New WebSocket connection: Client %s joined Room %s", clientID, roomID)
}