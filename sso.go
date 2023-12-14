package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type SingleSignOnToken struct {
	Id        string `json:"id"`
	Service   string `json:"service"`
	CreatedAt int64  `json:"createdAt"`
	ExpireAt  int64  `json:"expireAt"`
}

type SingleSignOnService struct {
	client   *http.Client
	service  string
	endpoint string
}

func NewSingleSignOnService() *SingleSignOnService {
	service := os.Getenv("SSO_SERVICE")
	if service == "" {
		service = "test"
	}
	endpoint := os.Getenv("SSO_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:13000"
	}
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	return &SingleSignOnService{service: service, endpoint: endpoint, client: client}
}

func (service *SingleSignOnService) ValidateToken(c *gin.Context) *SingleSignOnToken {
	tokenId, err := c.Cookie("token")
	if err != nil {
		return nil
	}
	token := service.GetToken(tokenId)
	if token == nil {
		return nil
	}
	now := time.Now().UnixMilli()
	if token.ExpireAt < now {
		return nil
	}
	return token
}

func (service *SingleSignOnService) GetToken(tokenId string) *SingleSignOnToken {
	endpoint := service.endpoint + "/api/token/" + service.service + "/" + tokenId
	resp, err := service.client.Get(endpoint)
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var token SingleSignOnToken
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil
	}
	return &token
}

func (service *SingleSignOnService) HandleTokenResp(c *gin.Context) {
	tokenId := c.Query("token")
	token := service.GetToken(tokenId)
	if token == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	now := time.Now().UnixMilli()
	if token.ExpireAt <= now {
		c.Status(http.StatusBadRequest)
		return
	}
	age := (token.ExpireAt - now) / 1000
	c.SetCookie("token", token.Id, int(age), "", "", false, false)
	c.Status(http.StatusOK)
}
