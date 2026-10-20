package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type IncrementPayload struct {
	ShortCode string `json:"short_code"`
}

const (
	TaskIncrement = "redirect_increment"
	TaskExpire    = "expire_link"
)

func IncrementLinkTask(shortCode string) (*asynq.Task, error) {
	body, marshalError := json.Marshal(IncrementPayload{ShortCode: shortCode})
	if marshalError != nil {
		return nil, marshalError
	}

	return asynq.NewTask(TaskIncrement, body), nil
}

func ExpireLinkTask() (*asynq.Task, error) {
	return asynq.NewTask(TaskExpire, nil, asynq.MaxRetry(5), asynq.Timeout(time.Second*20)), nil
}

func NewClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}
