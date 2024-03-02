package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type LuaResponse struct {
	LuaResponse string `json:"luaResponse"`
}

type Callbacks struct {
}

func NewCallbacks() *Callbacks {
	return &Callbacks{}
}

func (c *Callbacks) CallLuaFunction(redisClient *redis.Client, args ...interface{}) func(string) ([]byte, error) {
	return func(key string) ([]byte, error) {
		result, err := redisClient.Do(context.Background(), args...).Result()
		if err != nil {
			fmt.Printf("error executing FCALL for Lua function: %s\n", err)
			return nil, err
		}

		response := LuaResponse{
			LuaResponse: fmt.Sprintf("%v", result),
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Printf("error marshalling JSON: %v\n", err)
			return nil, err
		}
		return jsonData, nil
	}
}

var CallbackMap = map[string]func(*Callbacks, *redis.Client, string) ([]byte, error){
	"SaveSubscriptionGroup": func(c *Callbacks, client *redis.Client, key string) ([]byte, error) {
		asubID, channelName := parseKey(key)
		return c.CallLuaFunction(client, "FCALL", "SaveSubscriptionGroup", 2, asubID, channelName)(key)
	},
}

func parseKey(key string) (string, string) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 {
		return "", ""
	}
	asubID := parts[1]
	channelName := parts[2]
	return asubID, channelName
}
