package internal

// Application type
type Application struct {
	Token string `uri:"token"`
}

// Chat type
type Chat struct {
	Number int `uri:"number"`
}

// Message type
type Message struct {
	ApplicationToken string `json:"application_token,omitempty"`
	Body             string `json:"body,omitempty"`
	Number           int    `json:"number,omitempty"`
}

// RabbitMQMessage type
type RabbitMQMessage struct {
	Model            string `json:"model"`
	Number           int    `json:"number"`
	ApplicationToken string `json:"application_token"`
	Timestamp        int64  `json:"timestamp"`
	MessageBody      string `json:"message_body,omitempty"`
	ChatNumber       int    `json:"chat_number"`
}

// QueryResult type
type QueryResult struct {
	Number int
}
