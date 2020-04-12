package utils

import (
	uuid "github.com/satori/go.uuid"
	"os"
	"testing"
)

func TestJoinNotifyUrl(t *testing.T) {
	userId := uuid.NewV4()

	u := JoinNotifyUrl("http://www.baidu.com", 11, nil)
	if u != "http://www.baidu.com/11" {
		t.Errorf("result url %s unexpected", u)
	}

	u = JoinNotifyUrl("http://www.baidu.com/", 11, nil)
	if u != "http://www.baidu.com/11" {
		t.Errorf("result url %s unexpected", u)
	}

	os.Setenv("APP_REST_HOST", "business.test")
	u = JoinNotifyUrl("", 11, nil)
	if u != "http://business.test/pay/notify/11" {
		t.Errorf("result url %s unexpected", u)
	}

	u = JoinNotifyUrl("", 11, &userId)
	if u != "http://business.test/pay/notify/11" {
		t.Errorf("result url %s unexpected", u)
	}

	u = JoinNotifyUrl("http://www.baidu.com", 22, &userId)
	if u != "http://www.baidu.com/22/"+userId.String() {
		t.Errorf("result url %s unexpected", u)
	}
}
