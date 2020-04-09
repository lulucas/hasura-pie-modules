package utils

import (
	"net/url"
	"testing"
)

func TestNewSignFunc(t *testing.T) {
	fn := NewSignFunc(SignOption{
		IgnoreKeys: []string{"error_code", "error_msg", "sign"},
		PostSignStrHook: func(s string) string {
			return s + "ofCCzAV<|k|$c4DTT4}t<5vu"
		},
	})
	sign, err := fn(url.Values{
		"mch_id":       {"18674362989"},
		"method":       {"pay"},
		"back_url":     {""},
		"notify_url":   {"http://bf.doeal.top/notify"},
		"limit_pay":    {"1"},
		"total_fee":    {"1"},
		"out_trade_no": {"1231sal692223422112"},
		"type":         {"WXCODE"},
		"body":         {"女装"},
		"client_ip":    {"123.2.2.111"},
	})
	if err != nil {
		t.Error(err)
		return
	}

	// back_url=&body=女装&client_ip=123.2.2.111&limit_pay=1&mch_id=18674362989&method=pay&notify_url=http://bf.doeal.top/notify&out_trade_no=1231sal692223422112&total_fee=1&type=WXCODEofCCzAV<|k|$c4DTT4}t<5vu
	if sign != "f8bc7c2ce75e65f0e4047e8f7610b631" {
		t.Errorf("Sign error")
	}
}
