package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
)

// Ganti dengan Google OAuth Client ID Anda
// Dapatkan dari: https://console.cloud.google.com/apis/credentials
const googleClientID = "994808628494-qmgmj3vmrl0evvmnnbd66mnvpooa3na4.apps.googleusercontent.com"

// AuthRequest adalah struct untuk menerima token dari frontend
type AuthRequest struct {
	Token string `json:"token"`
}

// UserData menyimpan informasi user dari Google
type UserData struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// AuthResponse adalah struct untuk mengembalikan response ke frontend
type AuthResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    UserData `json:"data,omitempty"`
	Token   string   `json:"token,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/google", handleGoogleAuth)

	// Serve frontend static files dari folder "frontend"
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/", fs)

	handler := corsMiddleware(mux)

	log.Printf("🚀 Server berjalan di http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// corsMiddleware menambahkan CORS headers agar frontend bisa mengakses backend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Izinkan semua origin untuk development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleGoogleAuth menerima ID Token dari frontend, memverifikasi, dan mengembalikan data user
func handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Hanya terima method POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AuthResponse{
			Status:  "error",
			Message: "Method not allowed. Gunakan POST.",
		})
		return
	}

	// Parse request body
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Status:  "error",
			Message: "Request body tidak valid: " + err.Error(),
		})
		return
	}
	defer r.Body.Close()

	// Validasi token tidak kosong
	if req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Status:  "error",
			Message: "Token tidak boleh kosong.",
		})
		return
	}

	// Verifikasi Google ID Token
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, req.Token, googleClientID)
	if err != nil {
		log.Printf("❌ Token verification gagal: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Status:  "error",
			Message: "Token tidak valid atau sudah expired.",
		})
		return
	}

	// Ekstrak informasi user dari token claims
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	log.Printf("✅ Login berhasil: %s (%s)", name, email)

	// Kirim response sukses dengan data user
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Status:  "success",
		Message: "Login berhasil!",
		Token:   req.Token, // Kembalikan token untuk disimpan di cookie
		Data: UserData{
			Email:   email,
			Name:    name,
			Picture: picture,
		},
	})
}
