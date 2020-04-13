package utils

import (
	uuid "github.com/satori/go.uuid"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

func JoinNotifyUrl(section, rawUrl string, channelId int32, userId *uuid.UUID) string {
	if rawUrl == "" {
		if strings.ToLower(os.Getenv("APP_TLS_ENABLED")) == "true" {
			rawUrl = "https://"
		} else {
			rawUrl = "http://"
		}
		rawUrl += strings.ToLower(os.Getenv("APP_REST_HOST")) + "/pay/notify"
	}
	u, _ := url.Parse(rawUrl)
	joins := []string{u.Path, section, strconv.Itoa(int(channelId))}
	if userId != nil {
		joins = append(joins, userId.String())
	} else {
		joins = append(joins, "")
	}
	u.Path = path.Join(joins...)
	return u.String()
}
