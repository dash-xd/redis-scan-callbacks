package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v9"
)

type LuaResponse struct {
	LuaResponse string `json:"luaResponse"`
}

type ScanResponse struct {
	Keys   []string `json:"keys"`
	Cursor uint64   `json:"cursor"`
}

type Callbacks struct {
	client *redis.Client
}

func LazyLoadRedis(client *redis.Client) (*Callbacks, error) {
	fmt.Println("LazyLoading Redis ...")
	if client == nil {
		fmt.Println("Initializing Redis client...")
		client = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URI"),
			Password: os.Getenv("REDISCLI_AUTH"),
			DB:       0,
		})

		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			return nil, fmt.Errorf("error initializing Redis client: %v", err)
		}
	}

	return &Callbacks{client: client}, nil
}

func (c *Callbacks) interpretScanResponse(keys []string, cursor uint64) ([]byte, error) {
	response := ScanResponse{
		Keys:   keys,
		Cursor: cursor,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %v", err)
	}

	return jsonData, nil
}

func (c *Callbacks) callLuaFunction(luaFunctionName string, args ...interface{}) func(*Callbacks, string) {
	return func(c *Callbacks, key string) {
		result, err := c.client.Do(context.Background(), "FCALL", luaFunctionName, args...).Result()
		if err != nil {
			fmt.Printf("error executing FCALL for Lua function %s: %s\n", luaFunctionName, err)
			return
		}

		response := LuaResponse{
			LuaResponse: fmt.Sprintf("%v", result),
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("error marshalling JSON: %v\n", err)
			return
		}
		return jsonData
	}
}

var callbackMap = map[string]func(*Callbacks, string){
	"interpretScanResponse":       (*Callbacks).interpretScanResponse,
	"RegisterActiveSubscription": (*Callbacks).callLuaFunction("RegisterActiveSubscription"),
	"SaveSubscriptionGroup":      (*Callbacks).callLuaFunction("SaveSubscriptionGroup"),
}
