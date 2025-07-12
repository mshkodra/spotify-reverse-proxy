package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	spotifyauth "golang.org/x/oauth2/spotify"
)

var oauthConfig *oauth2.Config

var tokenStore = struct {
	sync.RWMutex
	m map[string]*oauth2.Token
}{m: make(map[string]*oauth2.Token)}

type SpotifyUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
	Country string `json:"country"`
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	oauthConfig = &oauth2.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifyClientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"user-read-private", "playlist-read-private",
		},
		Endpoint: spotifyauth.Endpoint,
	}

	if spotifyClientID == "" || spotifyClientSecret == "" || redirectURI == "" {
		log.Fatal("Missing required environment variables")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Spotify Reverse Proxy is running!\n")
	})

	http.HandleFunc("/login", handleLogin)

	http.HandleFunc("/callback", handleCallback)

	fmt.Printf("Server started on http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("Redirecting user to:", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), code)

	if err != nil {
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		http.Error(w, "Failed to fetch user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var user SpotifyUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(w, "Failed to unmarshal user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenStore.Lock()
	tokenStore.m[user.ID] = token
	tokenStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:  "user_id",
		Value: user.ID,
		Path:  "/",
	})

}
