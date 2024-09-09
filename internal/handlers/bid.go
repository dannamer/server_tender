package handlers

// import (
// 	"encoding/json"
// 	"net/http"
// 	"tender-service/internal/models"
// )

// func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
// 	var tender models.Tender
// 	err := json.NewDecoder(r.Body).Decode(&tender)
// 	if err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	// Логика сохранения тендера в базе данных
// 	// ...

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(tender)
// }

// func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
// 	var bid models.Bid
// 	err := json.NewDecoder(r.Body).Decode(&bid)
// 	if err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	bid.CreatedAt = time.Now()
// 	bid.UpdatedAt = time.Now()

// 	// Логика сохранения предложения в базу данных
// 	// Пример: database.SaveBid(bid)

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(bid)
// }
