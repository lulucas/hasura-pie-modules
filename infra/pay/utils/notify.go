package utils

import (
	pie "github.com/lulucas/hasura-pie"
	uuid "github.com/satori/go.uuid"
	"net/url"
	"path"
	"strconv"
)

func JoinNotifyUrl(section, rawUrl string, channelId int32, userId *uuid.UUID) string {
	if rawUrl == "" {
		rawUrl = pie.RestBaseUrl() + "/pay/notify"
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
