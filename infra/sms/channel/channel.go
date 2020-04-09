package channel

import (
	"encoding/json"
	"time"
)

type Channel interface {
	SendCaptcha(params json.RawMessage, mobile string, captcha string, ttl time.Duration) error
}
