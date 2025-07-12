package spotify

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

func ProxyHandler(tokenStore *TokenStore, oauthConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, _ := r.Cookie("user_id")
		token, ok := tokenStore.Get(userCookie.Value)

		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/spotify")
		req, _ := http.NewRequest("GET", "https://api.spotify.com/v1"+path, nil)
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch from Spotify: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
