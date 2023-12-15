package sso

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	tokenLen          int64 = 16
	tokenDurationInMs int64 = 30 * 24 * 3600 * 1000
)

type Token struct {
	Id        string `json:"id"`
	Service   string `json:"service"`
	CreatedAt int64  `json:"createdAt"`
	ExpireAt  int64  `json:"expireAt"`
}

type createTokenReq struct {
	result  chan *Token
	service string
}

type getTokenReq struct {
	result  chan *Token
	id      string
	service string
}

type deleteTokenReq struct {
	id      string
	service string
}

type tokenConfig struct {
	Service  string `json:"service"`
	Callback string `json:"callback"`
}

type TokenService struct {
	eventCh      chan interface{}
	serviceCofig []tokenConfig
	mappedTokens map[string]*Token
	sortedTokens []*Token
}

func NewTokenService() *TokenService {
	config, err := os.Open("service.json")
	if err != nil {
		return nil
	}
	defer config.Close()
	var serviceConfig []tokenConfig
	err = json.NewDecoder(config).Decode(&serviceConfig)
	if err != nil {
		return nil
	}
	service := &TokenService{
		make(chan interface{}),
		serviceConfig,
		make(map[string]*Token),
		[]*Token{},
	}
	go service.Run()
	return service
}

func newTokenId() (string, error) {
	bytes := make([]byte, tokenLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (service *TokenService) purgeExpiredTokens(now int64) {
	i := 0
	for i < len(service.sortedTokens) && service.sortedTokens[i].ExpireAt <= now {
		delete(service.mappedTokens, service.sortedTokens[i].Id)
		i++
	}
	if i > 0 {
		service.sortedTokens = service.sortedTokens[i:]
	}
}

func (service *TokenService) RequestToken(requestedService string) *Token {
	result := make(chan *Token)
	req := &createTokenReq{result: result, service: requestedService}
	service.eventCh <- req
	return <-result
}

func (service *TokenService) GetToken(id string, requestedService string) *Token {
	result := make(chan *Token)
	req := &getTokenReq{id: id, service: requestedService, result: result}
	service.eventCh <- req
	return <-result
}

func (service *TokenService) DeleteToken(id string, requestedService string) {
	req := &deleteTokenReq{id: id, service: requestedService}
	service.eventCh <- req
}

func (service *TokenService) GetServiceCallback(token, requestedService string) string {
	for _, config := range service.serviceCofig {
		if config.Service == requestedService {
			return fmt.Sprintf(config.Callback, token)
		}
	}
	return ""
}

func (service *TokenService) handleRequestToken(now int64, requestedService string) (*Token, error) {
	tokenId, err := newTokenId()
	for {
		if err != nil {
			return nil, err
		}
		if _, exists := service.mappedTokens[tokenId]; !exists {
			break
		}
		tokenId, err = newTokenId()
	}
	newToken := &Token{tokenId, requestedService, now, now + tokenDurationInMs}
	service.mappedTokens[tokenId] = newToken
	service.sortedTokens = append(service.sortedTokens, newToken)
	return newToken, nil
}

func (service *TokenService) handleGetToken(now int64, id, requestedService string) *Token {
	token, ok := service.mappedTokens[id]
	if !ok {
		return nil
	}
	if token.Service != requestedService {
		return nil
	}
	if token.ExpireAt <= now {
		return nil
	}
	return token
}

func (service *TokenService) handleDeleteToken(id, requestedService string) {
	token, ok := service.mappedTokens[id]
	if !ok {
		return
	}
	if token.Service != requestedService {
		return
	}
	token.ExpireAt = 0
}

func (service *TokenService) Run() {
	for event := range service.eventCh {
		now := time.Now().UnixMilli()
		service.purgeExpiredTokens(now)
		switch req := event.(type) {
		case *createTokenReq:
			token, _ := service.handleRequestToken(now, req.service)
			req.result <- token
		case *getTokenReq:
			token := service.handleGetToken(now, req.id, req.service)
			req.result <- token
		case *deleteTokenReq:
			service.handleDeleteToken(req.id, req.service)
		}
	}
}
