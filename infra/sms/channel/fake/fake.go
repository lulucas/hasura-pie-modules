package fake

import (
	"encoding/json"
	"time"
)

type Fake struct {
}

func New() *Fake {
	return &Fake{}
}

func (f *Fake) SendCaptcha(params json.RawMessage, mobile string, captcha string, ttl time.Duration) error {
	return nil
}

type Params struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}
