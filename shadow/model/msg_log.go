package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/shadow/app"
)

type MsgLogFilter struct {
	Pk        string // 产品标识
	Sn        string // 设备SN
	LogTypes  []string
	Topics    []string //过滤主题
	MsgId     string   //消息id
	ContextId string   //上下文id
	Dir       string
	Code      string
}

type MsgLog struct {
	Pk        string    `json:"pk"`         // 产品标识
	Sn        string    `json:"sn"`         // 设备SN
	Content   string    `json:"content"`    // 具体信息
	Topic     string    `json:"topic"`      // 主题 mqtt
	LogType   string    `json:"log_type"`   // 操作类型
	Dir       string    `json:"dir"`        // 消息方向
	Ts        time.Time `json:"ts"`         // 操作时间
	MsgId     string    `json:"msg_id"`     // 消息Id
	ContextId string    `json:"context_id"` // 上下文Id
	Result    string    `json:"result"`     // 结果
	Code      string    `json:"code"`       // 消息代码

}

func arrayToSql[arrType any](arr []arrType) (sql string) {
	if len(arr) == 0 {
		return ""
	}
	for _, v := range arr {
		sql += fmt.Sprintf("\"%v\",", v)
	}
	sql = sql[:len(sql)-1]
	return
}

func getLogStableName() string {
	return fmt.Sprintf("`msg_log`")
}

func getDeviceTableName(pk, sn string) string {
	return fmt.Sprintf("`device_msg_log_%s_%s`", pk, sn)
}

func CreateDeviceMsgLogStable(ctx context.Context) error {
	sql := fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s "+
		"(`ts` timestamp,`content` BINARY(1024),`topic` BINARY(128), `log_type` BINARY(64),`dir` BINARY(32), `msg_id` BINARY(64), `context_id` BINARY(64), `result` BINARY(1024), `code` BINARY(512)) "+
		"TAGS (`pk` BINARY(50),`sn` BINARY(50));", getLogStableName())
	if _, err := app.TD.ExecContext(ctx, sql); err != nil {
		panic(err)
	}

	return nil
}

func MsgLogFillFilter(sql sq.SelectBuilder, filter *MsgLogFilter) sq.SelectBuilder {
	if len(filter.Pk) != 0 {
		sql = sql.Where("`pk`=\"?\"", filter.Pk)
	}
	if len(filter.Sn) != 0 {
		sql = sql.Where("`sn`=\"?\"", filter.Sn)
	}
	if len(filter.Dir) != 0 {
		sql = sql.Where("`dir`=\"?\"", filter.Dir)
	}
	if len(filter.MsgId) != 0 {
		sql = sql.Where("`msg_id`=\"?\"", filter.MsgId)
	}
	if len(filter.ContextId) != 0 {
		sql = sql.Where("`context_id`=\"?\"", filter.ContextId)
	}
	if len(filter.Code) != 0 {
		sql = sql.Where("`code`=\"?\"", filter.Code)
	}
	if len(filter.LogTypes) != 0 {
		sql = sql.Where(fmt.Sprintf("`log_type` in (%v)", arrayToSql(filter.LogTypes)))
	}
	if len(filter.Topics) != 0 {
		sql = sql.Where(fmt.Sprintf("`topic` in (%v)", arrayToSql(filter.Topics)))
	}
	return sql
}

func MsgLogCount(ctx context.Context, filter *MsgLogFilter, page *PageInfo) (int64, error) {
	sqSql := sq.Select("Count(1)").From(getLogStableName())
	sqSql = MsgLogFillFilter(sqSql, filter)
	sqSql = page.FmtWhere(sqSql)
	sqlStr, value, err := sqSql.ToSql()
	logger.Infof("sqlStr:%s, value:%s", sqlStr, value)
	if err != nil {
		return 0, err
	}
	row := app.TD.QueryRowContext(ctx, sqlStr, value...)
	if err != nil {
		return 0, err
	}
	var (
		total int64
	)

	err = row.Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}

func MsgLogPage(ctx context.Context, filter *MsgLogFilter, page *PageInfo) (
	[]*MsgLog, error) {
	sql := sq.Select("*").From(getLogStableName()).OrderBy("`ts` desc")
	sql = MsgLogFillFilter(sql, filter)
	sql = page.FmtSql(sql)
	sqlStr, value, err := sql.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := app.TD.QueryContext(ctx, sqlStr, value...)
	if err != nil {
		return nil, err
	}

	retLogs := make([]*MsgLog, 0, 0)
	defer rows.Close()
	for rows.Next() {
		var r MsgLog
		err := rows.Scan(&r.Ts, &r.Content, &r.Topic, &r.LogType, &r.Dir, &r.MsgId, &r.ContextId, &r.Result, &r.Code, &r.Pk, &r.Sn)
		if err != nil {
			log.Fatalln("scan error:\n", err)
			return nil, err
		}
		retLogs = append(retLogs, &r)
	}

	return retLogs, nil
}

func MsgLogInsert(ctx context.Context, data *MsgLog) error {
	sql := fmt.Sprintf("insert into %s using %s tags('%s','%s')(`ts`,`content`,`topic`,"+
		"`log_type`,`dir`,`msg_id`,`context_id`,`result`,`code`) values ('%s','%s','%s','%s','%s','%s','%s','%s','%s');",
		getDeviceTableName(data.Pk, data.Sn), getLogStableName(), data.Pk, data.Sn, data.Ts.Format("2006-01-02 15:04:05.000"), data.Content, data.Topic, data.LogType, data.Dir, data.MsgId, data.ContextId, data.Result, data.Code)
	if _, err := app.TD.ExecContext(ctx, sql); err != nil {
		return err
	}
	return nil
}
