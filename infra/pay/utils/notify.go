package utils

import (
	uuid "github.com/satori/go.uuid"
	"net/url"
	"path"
	"strconv"
)

// 通道编号和用户编号写入路径
func JoinNotifyUrl(rawUrl string, channelId int32, userId *uuid.UUID) string {
	u, _ := url.Parse(rawUrl)
	joins := []string{u.Path, strconv.Itoa(int(channelId))}
	if userId != nil {
		joins = append(joins, userId.String())
	}
	u.Path = path.Join(joins...)
	return u.String()
}
