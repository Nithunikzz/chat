package api

import (
	"chatroom/chat"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HandleJoin handles the /join endpoint
func HandleJoin(chatRoom *chat.ChatRoom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("id")
		if clientID == "" {
			http.Error(w, "Missing client ID", http.StatusBadRequest)
			return
		}
		chatRoom.RegisterClient(clientID)
		w.WriteHeader(http.StatusOK)
	}
}

// HandleLeave handles the /leave endpoint
func HandleLeave(chatRoom *chat.ChatRoom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("id")
		if clientID == "" {
			http.Error(w, "Missing client ID", http.StatusBadRequest)
			return
		}
		chatRoom.UnregisterClient(clientID)
		w.WriteHeader(http.StatusOK)
	}
}

// HandleSend handles the /send endpoint
func HandleSend(chatRoom *chat.ChatRoom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("id")
		message := r.URL.Query().Get("message")
		if clientID == "" || message == "" {
			http.Error(w, "Missing client ID or message", http.StatusBadRequest)
			return
		}
		chatRoom.BroadcastMessage(clientID, message)
		w.WriteHeader(http.StatusOK)
	}
}

// HandleGetMessages handles the /messages endpoint
func HandleGetMessages(chatRoom *chat.ChatRoom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("id")
		if clientID == "" {
			http.Error(w, "Missing client ID", http.StatusBadRequest)
			return
		}
		message, ok := chatRoom.GetMessage(clientID, 10*time.Second)
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	}
}

// HandleFetchHistory handles the /history endpoint
func HandleFetchHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		history, err := db.Query("SELECT sender_id, message, created_at FROM messages ORDER BY created_at ASC")
		if err != nil {
			http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
			return
		}
		defer history.Close()

		var messages []map[string]string
		for history.Next() {
			var sender, message, createdAt string
			if err := history.Scan(&sender, &message, &createdAt); err != nil {
				http.Error(w, "Failed to scan message", http.StatusInternalServerError)
				return
			}
			messages = append(messages, map[string]string{
				"sender":     sender,
				"message":    message,
				"created_at": createdAt,
			})
		}
		json.NewEncoder(w).Encode(messages)
	}
}
