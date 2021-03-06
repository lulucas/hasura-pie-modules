package pay

import (
	"github.com/go-pg/pg/v9"
	"github.com/lulucas/hasura-pie-modules/infra/pay/model"
	"github.com/lulucas/hasura-pie-modules/infra/pay/utils"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrPayChannelNotFound  = errors.New("pay.channel-not-found")
	ErrPayChannelNoEnabled = errors.New("pay.channel-not-enabled")
)

func (m *pay) Pay(section string, channelId int32, orderId uuid.UUID, userId *uuid.UUID, amount decimal.Decimal,
	subject, body, returnUrl, clientIp string) (method string, data string, err error) {

	payChannel := model.PayChannel{}
	if err := m.db.Model(&payChannel).Where("id = ?", channelId).Select(); err != nil {
		if err == pg.ErrNoRows {
			return "", "", ErrPayChannelNotFound
		}
		return "", "", err
	}

	if !payChannel.Enabled {
		return "", "", ErrPayChannelNoEnabled
	}

	if returnUrl == "" {
		returnUrl = payChannel.ReturnUrl
	}

	// find channel handler
	ch, ok := m.channels[payChannel.Platform]
	if !ok {
		return "", "", ErrPayChannelNotFound
	}

	notifyUrl := utils.JoinNotifyUrl(section, payChannel.NotifyUrl, channelId, userId)

	return ch.Pay(orderId, amount, subject, body, returnUrl, notifyUrl, clientIp, payChannel.Params)
}
