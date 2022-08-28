package creator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

var ctx = context.Background()

func GetAccessToken() string {

	//redisClient := redis.NewRedis()
	//token := redisClient.Redis().Get(ctx, "zoho-token-calc").Val()
	token, _ := refreshAuthToken()

	return token
}

/*
	Refresh Access Token Data
*/
func refreshAuthToken() (string, error) {

	response, err := http.Post("https://accounts.zoho.com/oauth/v2/token?refresh_token="+os.Getenv("REFRESH_TOKEN")+"&"+
		"client_secret="+os.Getenv("CLIENT_SECRET")+"&client_id="+os.Getenv("CLIENT_ID")+
		"&grant_type=refresh_token", "application/json", nil)

	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 200 {
		err := errors.New("Invalid response on token refresh request. ")
		return "", err
	}
	var res map[string]interface{}

	json.NewDecoder(response.Body).Decode(&res)

	accessToken := fmt.Sprintf("%v", res["access_token"])
	return accessToken, nil
}
