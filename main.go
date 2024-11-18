package main

import (
	"chatroom/api"
	"chatroom/chat"
	"chatroom/db"
	"log"
	"net/http"
)

func main() {
	// Initialize the database
	database, err := db.InitDB("./chat.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize the chat room
	chatRoom := chat.NewChatRoom(database)
	go chatRoom.Run()

	// Set up routes
	http.HandleFunc("/join", api.HandleJoin(chatRoom))
	http.HandleFunc("/leave", api.HandleLeave(chatRoom))
	http.HandleFunc("/send", api.HandleSend(chatRoom))
	http.HandleFunc("/messages", api.HandleGetMessages(chatRoom))
	http.HandleFunc("/history", api.HandleFetchHistory(database))

	// Start the server
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
