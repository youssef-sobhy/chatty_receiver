package internal

import (
	"database/sql"
	"fmt"
	"os"
)

const (
	// MessageNumberQuery const
	MessageNumberQuery = `SELECT messages.number
		FROM messages
		INNER JOIN chats ON chats.id = messages.chat_id
		INNER JOIN applications ON applications.id = chats.application_id
		WHERE applications.token = ?
		AND chats.number = ?
		ORDER BY messages.created_at DESC
		LIMIT 1`

	// ChatNumberQuery const
	ChatNumberQuery = `SELECT number
		FROM chats
		INNER JOIN applications ON applications.id = chats.application_id
		WHERE applications.token = ?
		ORDER BY chats.created_at DESC
		LIMIT 1`
)

var mysqlURL = os.Getenv("MYSQL_URL")

// GetNumber func
func GetNumber(query string, token string, chatNumber int) int {
	// Connect to DB
	db, err := sql.Open("mysql", mysqlURL)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var queryResult QueryResult

	err = db.QueryRow(query, token, chatNumber).Scan(&queryResult.Number)
	if queryResult.Number == 0 {
		return 1
	}
	return queryResult.Number + 1
}
