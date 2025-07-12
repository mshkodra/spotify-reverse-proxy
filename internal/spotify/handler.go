package spotify

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

func forwardSpotifyRequest(accessToken, path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1"+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	return client.Do(req)
}

func ProxyHandler(tokenStore *RedisTokenStore, oauthConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, "Missing user ID cookie", http.StatusUnauthorized)
			return
		}

		token, err := tokenStore.Get(userCookie.Value)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/spotify")

		resp, err := forwardSpotifyRequest(token.AccessToken, path)
		if err != nil {
			http.Error(w, "Failed to fetch from Spotify: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			newToken, err := oauthConfig.TokenSource(r.Context(), token).Token()
			if err != nil {
				http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
				return
			}

			err = tokenStore.Set(userCookie.Value, newToken)
			if err != nil {
				http.Error(w, "Failed to update token in store: "+err.Error(), http.StatusInternalServerError)
				return
			}

			resp.Body.Close()
			resp, err = forwardSpotifyRequest(newToken.AccessToken, path)
			if err != nil {
				http.Error(w, "Failed to retry after token refresh: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
		}

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
