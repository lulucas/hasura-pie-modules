package pay

import (
	"github.com/labstack/echo/v4"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/lulucas/hasura-pie-modules/infra/pay/model"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
)

func notify(cc pie.CreatedContext, channels map[string]channel.Channel) echo.HandlerFunc {
	return func(c echo.Context) error {
		channelIdStr := c.Param("id")
		channelId, err := strconv.Atoi(channelIdStr)
		if err != nil {
			return err
		}
		userIdStr := c.Param("uid")
		var userId *uuid.UUID
		if userIdStr != "" {
			id, err := uuid.FromString(userIdStr)
			if err != nil {
				return err
			}
			userId = &id
		}

		tx, err := cc.DB().WithContext(c.Request().Context()).Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		payChannel := model.PayChannel{}
		if err := tx.Model(&payChannel).Where("id = ?", channelId).Select(); err != nil {
			return err
		}

		cc.Logger().Infof("Notify pay channel id: %d, name: %s", payChannel.Id, payChannel.Platform)

		ch, ok := channels[payChannel.Platform]
		if !ok {
			cc.Logger().Errorf("Notify pay channel id: %s not found", channelId)
			return c.String(http.StatusBadRequest, "pay channel not found")
		}

		notification, err := ch.Notify(c, payChannel.Params)
		if err != nil {
			cc.Logger().Errorf("Notify validation error: %s", err.Error())
			return err
		}

		payLog := model.PayLog{
			PayChannelId:   int32(channelId),
			UserId:         userId,
			OrderId:        notification.OrderId,
			OutTradeNo:     notification.OutTradeNo,
			OrderAmount:    notification.OrderAmount,
			ReceivedAmount: notification.ReceivedAmount,
			Attach:         notification.Attach,
			IsPaid:         notification.IsPaid,
		}
		if err := tx.Insert(&payLog); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		cc.Logger().Infof("Create pay log for order %s through channel id: %d, name: %s",
			notification.OrderId, channelId, payChannel.Platform)

		if err := ch.ConfirmNotify(c); err != nil {
			return err
		}

		cc.Logger().Infof("Confirm pay notify id: %d", channelId)

		return nil
	}
}
