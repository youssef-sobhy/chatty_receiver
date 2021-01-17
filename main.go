package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/streadway/amqp"
)

const (
	rabbitMQQueueName = "receiver.messages"
)

var (
	rabbitMQURL   = os.Getenv("RABBITMQ_URL")
	mysqlURL      = os.Getenv("MYSQL_URL")
	redisAddr     = os.Getenv("REDIS_ADDR")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	redisDB       = os.Getenv("REDIS_DB")
	ctx           = context.Background()
)

// Application token binding
type Application struct {
	Token string `uri:"token"`
}

// Chat number binding
type Chat struct {
	Number int `uri:"number"`
}

// Message binding
type Message struct {
	ApplicationToken string `json:"application_token,omitempty"`
	Body             string `json:"body,omitempty"`
	Number           int    `json:"number,omitempty"`
}

// rabbitMQMessage struct
type rabbitMQMessage struct {
	Model            string `json:"model"`
	Number           int    `json:"number"`
	ApplicationToken string `json:"application_token"`
	Timestamp        int64  `json:"timestamp"`
	MessageBody      string `json:"message_body,omitempty"`
	ChatNumber       int    `json:"chat_number"`
}

func main() {
	r := gin.Default()
	// Connect to Redis
	redisDB, _ := strconv.Atoi(redisDB)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	createdAt := time.Now()
	timestamp := createdAt.UnixNano() / int64(time.Millisecond)
	var number int

	r.POST("api/v1/applications/:token/chats", func(c *gin.Context) {
		var application Application
		c.ShouldBindUri(&application)
		lastChatNumber, err := rdb.Get(ctx, application.Token).Result()
		if err != nil {
			number = getChatNumber(application.Token)
		} else {
			lastChatNumber, _ := strconv.Atoi(lastChatNumber)
			number = lastChatNumber + 1
		}
		rdb.Set(ctx, application.Token, number, 0)
		rabbitMQMessage, _ := json.Marshal(&rabbitMQMessage{"Chat", number, application.Token, timestamp, "", 0})
		if err := publish(rabbitMQMessage); err != nil {
			fmt.Println(err.Error())
		}
		c.JSON(200, gin.H{"application_token": application.Token, "number": number, "created_at": createdAt})
	})

	r.POST("api/v1/chats/:number/messages", func(c *gin.Context) {
		var chat Chat
		message := new(Message)
		c.ShouldBindUri(&chat)
		c.Bind(message)
		key := message.ApplicationToken + "-" + strconv.Itoa(chat.Number)
		lastMessageNumber, err := rdb.Get(ctx, key).Result()
		if err != nil {
			number = getMessageNumber(message.ApplicationToken, chat.Number)
		} else {
			lastMessageNumber, _ := strconv.Atoi(lastMessageNumber)
			number = lastMessageNumber + 1
		}
		rdb.Set(ctx, key, number, 0)
		rabbitMQMessage, _ := json.Marshal(rabbitMQMessage{"Message", number, message.ApplicationToken, timestamp, message.Body, chat.Number})
		if err := publish(rabbitMQMessage); err != nil {
			fmt.Println(err.Error())
		}
		c.JSON(200, gin.H{"application_token": message.ApplicationToken, "number": number, "chat_number": chat.Number, "created_at": createdAt})
	})
	r.Run(":3001")
}

func publish(m []byte) error {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// init RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Init RabbitMQ queue
	args := make(amqp.Table)
	args["x-dead-letter-exchange"] = "receiver.messages-retry"

	q, err := ch.QueueDeclare(
		rabbitMQQueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments
	)
	if err != nil {
		return err
	}

	// Publish to RabbitMQ
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         m,
		})
	if err != nil {
		return err
	}

	return nil
}

func getChatNumber(token string) int {
	// Connect to DB
	db, err := sql.Open("mysql", mysqlURL)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var chat Chat

	err = db.QueryRow(`SELECT number
		FROM chats
		INNER JOIN applications ON applications.id = chats.application_id
		WHERE applications.token = ?
		ORDER BY chats.created_at DESC
		LIMIT 1`, token).Scan(&chat.Number)
	if chat.Number == 0 {
		return 1
	}
	return chat.Number + 1
}

func getMessageNumber(token string, chatNumber int) int {
	// Connect to DB
	db, err := sql.Open("mysql", mysqlURL)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var message Message

	err = db.QueryRow(`SELECT messages.number
		FROM messages
		INNER JOIN chats ON chats.id = messages.chat_id
		INNER JOIN applications ON applications.id = chats.application_id
		WHERE applications.token = ?
		AND chats.number = ?
		ORDER BY messages.created_at DESC
		LIMIT 1`, token, chatNumber).Scan(&message.Number)
	if message.Number == 0 {
		return 1
	}
	return message.Number + 1
}
