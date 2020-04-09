package utils

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestJoinNotifyUrl(t *testing.T) {
	u := JoinNotifyUrl("http://www.baidu.com", 11, nil)
	if u != "http://www.baidu.com/11" {
		t.Errorf("result url unexpected")
	}

	u = JoinNotifyUrl("http://www.baidu.com/", 11, nil)
	if u != "http://www.baidu.com/11" {
		t.Errorf("result url unexpected")
	}

	userId := uuid.NewV4()
	u = JoinNotifyUrl("http://www.baidu.com", 22, &userId)
	if u != "http://www.baidu.com/22/"+userId.String() {
		t.Errorf("result url unexpected")
	}
}
