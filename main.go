package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/chatty_receiver/internal"
)

func main() {
	r := gin.Default()

	createdAt := time.Now()
	timestamp := createdAt.UnixNano() / int64(time.Millisecond)

	// Chat creation endpoint
	r.POST("api/v1/applications/:token/chats", func(c *gin.Context) {
		var application internal.Application
		c.ShouldBindUri(&application)
		number := internal.CachedNumber("Chat", application.Token, 0)

		fmt.Println(number)

		rabbitMQMessage, _ := json.Marshal(internal.RabbitMQMessage{
			Model:            "Chat",
			Number:           number,
			ApplicationToken: application.Token,
			Timestamp:        timestamp,
			MessageBody:      "",
			ChatNumber:       0})

		if err := internal.Publish(rabbitMQMessage); err != nil {
			fmt.Println(err.Error())
		}

		c.JSON(200, gin.H{
			"application_token": application.Token,
			"number":            number,
			"created_at":        createdAt})
	})

	// Message creation endpoint
	r.POST("api/v1/chats/:number/messages", func(c *gin.Context) {
		var chat internal.Chat
		message := new(internal.Message)
		c.ShouldBindUri(&chat)
		c.Bind(message)
		number := internal.CachedNumber("Message", message.ApplicationToken, chat.Number)

		rabbitMQMessage, _ := json.Marshal(internal.RabbitMQMessage{
			Model:            "Message",
			Number:           number,
			ApplicationToken: message.ApplicationToken,
			Timestamp:        timestamp,
			MessageBody:      message.Body,
			ChatNumber:       chat.Number})

		if err := internal.Publish(rabbitMQMessage); err != nil {
			fmt.Println(err.Error())
		}

		c.JSON(200, gin.H{
			"application_token": message.ApplicationToken,
			"number":            number,
			"chat_number":       chat.Number,
			"created_at":        createdAt})
	})
	r.Run(":3001")
}
