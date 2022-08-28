package creator

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

var lock = &sync.Mutex{}

var instance *Single

type Single struct {
	httpClient http.Client
	httpQuery  map[string]string
	baseUrl    string
	token      string
}

func GetInstance() *Single {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		instance = &Single{token: GetAccessToken()}
	}

	return instance
}

/*
	Access Token
*/
func (s *Single) getAuthToken() string {
	accessToken := GetAccessToken()

	return accessToken
}

/*
	Get Base Url Value
*/
func (s *Single) GetBaseUrl() string {
	return s.baseUrl
}

/*
	Set Http Query
*/
func (s *Single) SetHttpQuery(key string, value string) {
	log.Println(s.httpQuery)
	m := make(map[string]string)
	m[key] = value
	s.httpQuery = m
}

// Request Api Method

func (s *Single) Request(method string, uri string, body io.Reader, query map[string]string) ([]byte, error) {
	params := url.Values{}
	for key, value := range query {
		params.Add(key, value)
	}

	request, _ := http.NewRequest(method, uri+"?"+params.Encode(), body)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Zoho-oauthtoken "+s.token)

	response, err := s.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	bodyData, _ := ioutil.ReadAll(response.Body)
	return bodyData, nil
}
