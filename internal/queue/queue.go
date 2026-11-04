package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type IncrementPayload struct {
	ShortCode string `json:"short_code"`
}

type RefreshHashPayload struct {
	RefreshHash string    `json:"refresh_hash"`
	Expiration  time.Time `json:"expiration"`
	UserID      string    `json:"user_id"`
}

type DeleteVerificationTokenByUserIDPayload struct {
	UserID string `json:"user_id"`
}

const (
	TaskIncrement                    = "redirect_increment"
	TaskExpire                       = "expire_link"
	TaskPersistRefresh               = "persist_refresh"
	TaskClearExpiredRefresh          = "clear_expired_refresh"
	TaskDeleteEmailVerificationToken = "delete_email_verification_token"
)

func IncrementLinkTask(shortCode string) (*asynq.Task, error) {
	body, marshalError := json.Marshal(IncrementPayload{ShortCode: shortCode})
	if marshalError != nil {
		return nil, marshalError
	}

	return asynq.NewTask(TaskIncrement, body), nil
}

func PersistRefreshTokenTask(refreshTokenHash string, expiration time.Time, userID string) (*asynq.Task, error) {
	body, marshalErr := json.Marshal(RefreshHashPayload{RefreshHash: refreshTokenHash, Expiration: expiration, UserID: userID})
	if marshalErr != nil {
		return nil, marshalErr
	}
	return asynq.NewTask(TaskPersistRefresh, body), nil
}

func ExpireLinkTask() *asynq.Task {
	return asynq.NewTask(TaskExpire, nil, asynq.MaxRetry(5), asynq.Timeout(time.Second*20))
}

func ClearExpiredRefreshHashTask() (*asynq.Task, error) {
	return asynq.NewTask(TaskClearExpiredRefresh, nil), nil
}

// DeleteEmailVerificationTokenByUserID builds an asynq task for async token cleanup.
// Verify/resend delete sync in the handlers today; this path is kept as an alternative.
func DeleteEmailVerificationTokenByUserID(userID string) (*asynq.Task, error) {
	var body, marshalError = json.Marshal(DeleteVerificationTokenByUserIDPayload{UserID: userID})
	if marshalError != nil {
		return nil, marshalError
	}

	return asynq.NewTask(TaskDeleteEmailVerificationToken, body), nil
}

func NewClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}
