package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

type LuaResponse struct {
	LuaResponse string `json:"luaResponse"`
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

func (c *Callbacks) callLuaFunction(args ...interface{}) func(string) ([]byte, error) {
    return func(key string) ([]byte, error) {
        result, err := c.client.Do(context.Background(), args...).Result()
        if err != nil {
            fmt.Printf("error executing FCALL for Lua function %s: %s\n", luaFunctionName, err)
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

var CallbackMap = map[string]func(*Callbacks, string) ([]byte, error){
    "SaveSubscriptionGroup": func(c *Callbacks, key string) ([]byte, error) {
        asubID, channelName := parseKey(key)
        return c.callLuaFunction("FCALL", "SaveSubscriptionGroup", 2, asubID, channelName)(key)
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

func parseKey(key string) (string, string) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 {
		return "", ""
	}
	asubID := parts[1]
	channelName := parts[2]   
	return asubID, channelName
}
