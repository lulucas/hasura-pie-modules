package analysis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/lulucas/hasura-pie-modules/analysis/model"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	visitKey       = "visit"
	systemClientId = "__sys__"
)

func hit(m *analysis) echo.HandlerFunc {
	return func(c echo.Context) error {
		gid := c.QueryParam("g")
		ref := c.QueryParam("r")
		u, err := url.Parse(c.Request().Referer())
		if err != nil {
			return err
		}
		cid := u.Host
		path := u.Path

		if err := m.Hit(systemClientId, gid, path, ref); err != nil {
			return err
		}
		if err := m.Hit(cid, gid, path, ref); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

func visitUniqueViews(m *analysis) interface{} {
	type VisitUniqueViewsOutput struct {
		UniqueViews []*model.DailyView
	}
	return func(ctx context.Context, input struct{
		ClientId *string
		StartTime *time.Time
		EndTime *time.Time
	}) (*VisitUniqueViewsOutput, error) {
		cid := systemClientId
		if input.ClientId != nil {
			cid = *input.ClientId
		}
		endTime := time.Now()
		startTime := endTime.AddDate(0, -1, 0)
		if input.StartTime != nil {
			startTime = *input.StartTime
		}
		if input.EndTime != nil {
			endTime = *input.EndTime
		}

		uvs, err := m.GetUniquesViews(cid, startTime, endTime)
		if err != nil {
			return nil, err
		}

		return &VisitUniqueViewsOutput{
			UniqueViews: uvs,
		}, nil
	}
}

func visitPageViews(m *analysis) interface{} {
	type VisitPageViewsOutput struct {
		PageViews []*model.DailyView
	}
	return func(ctx context.Context, input struct{
		ClientId *string
		StartTime *time.Time
		EndTime *time.Time
	}) (*VisitPageViewsOutput, error) {
		cid := systemClientId
		if input.ClientId != nil {
			cid = *input.ClientId
		}
		endTime := time.Now()
		startTime := endTime.AddDate(0, -1, 0)
		if input.StartTime != nil {
			startTime = *input.StartTime
		}
		if input.EndTime != nil {
			endTime = *input.EndTime
		}

		pvs, err := m.GetPageViews(cid, startTime, endTime)
		if err != nil {
			return nil, err
		}

		return &VisitPageViewsOutput{
			PageViews: pvs,
		}, nil
	}
}

func (m *analysis) Hit(cid, gid, path, ref string) error {
	pvKey := fmt.Sprintf("%s:%s:pvidx", visitKey, cid)
	pvIndex, err := m.r.Incr(pvKey).Result()
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	t := time.Now()
	unixTimestamp := float64(t.Unix())

	pvHashKey := fmt.Sprintf("%s:%s:pv:%d", visitKey, cid, pvIndex)
	pvPathKey := fmt.Sprintf("%s:%s:path:%s", visitKey, cid, path)
	pvRefKey := fmt.Sprintf("%s:%s:ref:%s", visitKey, cid, ref)
	timeIndexKey := fmt.Sprintf("%s:%s:timeidx", visitKey, cid)
	uvKey := fmt.Sprintf("%s:%s:uv:%s", visitKey, cid, t.In(loc).Format("2006-01-02"))

	pipe := m.r.Pipeline()
	pipe.HMSet(pvHashKey, map[string]interface{}{
		"gid": gid, "path": path, "ref": ref,
	})
	pipe.ZAdd(timeIndexKey, &redis.Z{
		Score:  unixTimestamp,
		Member: pvIndex,
	})
	pipe.ZAdd(pvPathKey, &redis.Z{
		Score:  unixTimestamp,
		Member: pvIndex,
	})
	pipe.ZAdd(pvRefKey, &redis.Z{
		Score:  unixTimestamp,
		Member: pvIndex,
	})
	pipe.PFAdd(uvKey, gid)

	if _, err := pipe.Exec(); err != nil {
		return err
	}
	return nil
}

// 获取UV
func (m *analysis) GetUniquesViews(cid string, start, end time.Time) ([]*model.DailyView, error) {
	pipe := m.r.Pipeline()

	type DateIntCmd struct {
		Date string
		Cmd  *redis.IntCmd
	}

	var dcs []*DateIntCmd
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.In(loc).Format("2006-01-02")
		dc := pipe.PFCount(fmt.Sprintf("%s:%s:uv:%s", visitKey, cid, date))
		dcs = append(dcs, &DateIntCmd{
			Date: date,
			Cmd:  dc,
		})
	}

	// 获取数量
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	var uvs []*model.DailyView
	for _, dc := range dcs {
		count, err := dc.Cmd.Result()
		if err != nil {
			return nil, err
		}
		uvs = append(uvs, &model.DailyView{
			Date:  dc.Date,
			Count: count,
		})
	}
	return uvs, nil
}

// 通过pattern扫描keys
func (m *analysis) getMatchingKeys(pattern string) []string {
	cur := uint64(0)

	var keys []string
	for {
		arr, cur, err := m.r.Scan(cur, pattern, 365).Result()
		if err != nil {
			return nil
		}
		keys = append(keys, arr...)
		if cur == 0 {
			break
		}
	}
	return keys
}

// 获取PV
func (m *analysis) GetPageViews(cid string, startTime, endTime time.Time) ([]*model.DailyView, error) {
	return m.getAllPageViews(cid, startTime, endTime)
}

// 通过path和ref获取PV
func (m *analysis) getPageViewsByRefAndPath(cid, path, ref string, startTime, endTime time.Time) int64 {
	pathKey := fmt.Sprintf("%s:%s:path:%s", visitKey, cid, path)
	refKey := fmt.Sprintf("%s:%s:ref:%s", visitKey, cid, ref)

	pv, err := m.r.ZInterStore(visitKey+":out", &redis.ZStore{
		Keys:    []string{pathKey, refKey},
		Weights: []float64{float64(startTime.Unix()), float64(endTime.Unix())},
	}).Result()
	if err != nil {
		return 0
	}
	return pv
}

// 通过ref获取PV
func (m *analysis) getPageViewsByRef(cid, ref string, startTime, endTime time.Time) int64 {
	refKey := fmt.Sprintf("%s:%s:ref:%s", visitKey, cid, ref)
	pv, err := m.r.ZCount(refKey, strconv.FormatInt(startTime.Unix(), 10), strconv.FormatInt(endTime.Unix(), 10)).Result()
	if err != nil {
		return 0
	}
	return pv
}

// 通过path获取PV
func (m *analysis) getPageViewsByPath(cid, ref string, startTime, endTime time.Time) int64 {
	refKey := fmt.Sprintf("%s:%s:path:%s", visitKey, cid, ref)
	pv, err := m.r.ZCount(refKey, strconv.FormatInt(startTime.Unix(), 10), strconv.FormatInt(endTime.Unix(), 10)).Result()
	if err != nil {
		return 0
	}
	return pv
}

// 获取所有PV
func (m *analysis) getAllPageViews(cid string, startTime, endTime time.Time) ([]*model.DailyView, error) {
	refKey := fmt.Sprintf("%s:%s:timeidx", visitKey, cid)

	type DateIntCmd struct {
		Date string
		Cmd  *redis.IntCmd
	}

	var dcs []*DateIntCmd
	pipe := m.r.Pipeline()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		cmd := pipe.ZCount(refKey, strconv.FormatInt(d.Unix(), 10), strconv.FormatInt(d.AddDate(0, 0, 1).Unix(), 10))
		date := d.In(loc).Format("2006-01-02")
		dcs = append(dcs, &DateIntCmd{
			Date: date,
			Cmd:  cmd,
		})
	}

	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	var pvs []*model.DailyView
	for _, dc := range dcs {
		count, err := dc.Cmd.Result()
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, &model.DailyView{
			Date:  dc.Date,
			Count: count,
		})
	}
	return pvs, nil
}
