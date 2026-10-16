package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

type IncrementPayload struct {
	ShortCode string `json:"short_code"`
}

const (
	TaskIncrement = "redirect_increment"
)

func IncrementLinkTask(shortCode string) (*asynq.Task, error) {
	body, marshalError := json.Marshal(IncrementPayload{ShortCode: shortCode})
	if marshalError != nil {
		return nil, marshalError
	}

	return asynq.NewTask(TaskIncrement, body), nil
}

func NewClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}
