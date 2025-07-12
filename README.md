# Reverse Proxy for Spotify API

First download repo on your computer.

Create a `.env` file and fill it with the following text:

```
SPOTIFY_REDIRECT_URI=http://localhost:8080/callback
SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
REDIS_ADDR=localhost:6379
```

Obviously fill in the `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET` with your respective API tokens.

Next run: `docker compose up --build` and server should be running at `localhost:8080`.

Go to `localhost:8080/login` to login to your Spotify account, and then after go to any `/spotify/...` endpoint you wish to go to.