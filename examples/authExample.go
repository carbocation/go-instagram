package main

import (
	"log"
	"net/http"
	"net/url"
)

func main() {
	code := "CODE"
	u := "redirect--------url"
	retrieveAccessToken(code, u)
}

func retrieveAccessToken(code string, redirectURL string) {
	urlStr := "https://api.instagram.com/oauth/access_token"
	params := url.Values{}
	params.Add("client_id", "CLIENT_ID")
	params.Add("client_secret", "CLIENT_SECRET")
	params.Add("grant_type", "authorization_code")
	params.Add("redirect_uri", redirectURL)
	params.Add("code", code)
	resp, err := http.PostForm(urlStr, params)
	if err != nil {
		log.Print("Error# ", err)
	}
	defer resp.Body.Close()
	log.Println("Resp =", resp)
}

/*
curl -F 'client_id=CLIENT_ID' \
    -F 'client_secret=CLIENT_SECRET' \
    -F 'grant_type=authorization_code' \
    -F 'redirect_uri=AUTHORIZATION_REDIRECT_URI' \
    -F 'code=CODE' \
    https://api.instagram.com/oauth/access_token
*/
