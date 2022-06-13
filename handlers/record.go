package handlers

import (
	"strings"
	"time"

	"github.com/azhai/bingwp/cmd"
	db "github.com/azhai/bingwp/models/default"

	"gitee.com/azhai/fiber-u8l/v2"
)

const (
	HEAD_KEY_UA = "UserAgent"
	MSG_URL_PRE = "/message/"
)

var (
	msgchan = make(chan map[string]any)
)

// MyErrorHandler 记录错误
func MyErrorHandler(ctx *fiber.Ctx) (err error) {
	return ctx.Abort(503, nil)
}

// MyGetHandler GET请求
func MyGetHandler(ctx *fiber.Ctx) (err error) {
	url := ctx.OriginalURL()
	ua, body := ctx.HeaderStr(HEAD_KEY_UA), ""
	go logRequest("GET", url, ua, body)
	err = ctx.Reply("success")
	return err
}

// MyPostHandler POST请求
func MyPostHandler(ctx *fiber.Ctx) (err error) {
	url := ctx.OriginalURL()
	ua, body := ctx.HeaderStr(HEAD_KEY_UA), ""
	ctype := ctx.ContentType().String()
	if ctype == "application/x-www-form-urlencoded" || ctype == "application/json" {
		body = string(ctx.Body())
	} else if bodyBytes := ctx.Body(); len(bodyBytes) < 2048 {
		body = string(bodyBytes)
	}
	go logRequest("POST", url, ua, body)

	if strings.HasPrefix(url, MSG_URL_PRE) {
		if data := parseMsgData(body); data != nil {
			msgchan <- data
		}
	}
	err = ctx.Reply("success")
	return err
}

// logRequest 记录请求到日志
func logRequest(act, url, ua, body string) {
	logger := cmd.GetDefaultLogger()
	if logger != nil {
		if body == "" {
			logger.Info(act, "\t", url, "\t", ua)
		} else {
			logger.Info(act, "\t", url, "\t", ua, "\t", body)
		}
	}
}

// logErrorIf 记录错误到日志
func logErrorIf(err error) {
	logger := cmd.GetDefaultLogger()
	if logger != nil && err != nil {
		logger.Error(err)
	}
}

// parseMsgData 解析消息内容
func parseMsgData(body string) (data map[string]any) {
	return
}

// SaveMsgData 保存到数据库
func SaveMsgData(writeSize int) {
	var rows []any
	// table := (db.MessageModel{}).TableName()
	table := "t_message"
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-tick:
			if len(rows) >= 0 { //每秒也入库
				logErrorIf(db.InsertBatch(table, rows))
				rows = nil
			}
		case data := <-msgchan:
			if data != nil {
				rows = append(rows, data)
			}
			if len(rows) >= writeSize { //超出长度也入库
				logErrorIf(db.InsertBatch(table, rows))
				rows = nil
			}
		}
	}
}
