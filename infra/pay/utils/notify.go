package utils

import (
	uuid "github.com/satori/go.uuid"
	"net/url"
	"os"
	"path"
	"strconv"
)

func JoinNotifyUrl(rawUrl string, channelId int32, userId *uuid.UUID) string {
	if rawUrl == "" {
		rawUrl = "https://" + os.Getenv("APP_REST_HOST")
	}
	u, _ := url.Parse(rawUrl)
	joins := []string{u.Path, strconv.Itoa(int(channelId))}
	if userId != nil {
		joins = append(joins, userId.String())
	}
	u.Path = path.Join(joins...)
	return u.String()
}
