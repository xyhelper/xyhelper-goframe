package xyhelper

import (
	"fmt"
	"net/http"
	"time"
	"xyhelper-goframe/modules/xyhelper/config"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/xyhelper/chatgpt-go"
)

// session
func Session(r *ghttp.Request) {
	type Data struct {
		Auth  bool   `json:"auth"`
		Model string `json:"model"`
	}
	type SessionRes struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    Data   `json:"data"`
	}
	r.Response.WriteJsonExit(&SessionRes{
		Status:  "Success",
		Message: "",
		Data: Data{
			Auth:  false,
			Model: "ChatGPTUnofficialProxyAPI",
		},
	})
}

// ChatProcessRequest
type ChatProcessRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Optins *struct {
		ConversationId  string `json:"conversationId"`  // 会话ID
		ParentMessageId string `json:"parentMessageId"` // 父消息ID
	} `json:"options"` // 选项
	BaseURI     string `json:"baseURI"`     // 基础URI
	AccessToken string `json:"accessToken"` // 访问令牌
	IsGPT4      bool   `json:"isGPT4"`      // 是否为GPT4
}

// ChatProcessResponse
type ChatProcessResponse struct {
	Role            string `json:"role"`            // 角色
	Id              string `json:"id"`              // 消息ID
	ParentMessageId string `json:"parentMessageId"` // 父消息ID
	ConversationId  string `json:"conversationId"`  // 会话ID
	Text            string `json:"text"`            // 消息内容
}

// ChatProcess
func ChatProcess(r *ghttp.Request) {
	var ctx = r.GetCtx()
	var req *ChatProcessRequest
	err := r.GetStruct(&req)
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(err)
	}
	g.DumpWithType(req)
	cli := chatgpt.NewClient(
		chatgpt.WithAccessToken(req.AccessToken),
		chatgpt.WithTimeout(time.Duration(config.TimeOutMs*1000*1000)),
		chatgpt.WithBaseURI(req.BaseURI),
	)
	if req.IsGPT4 {
		cli.SetModel("gpt-4")
	}
	stream, err := cli.GetChatStream(req.Prompt, req.Optins.ConversationId, req.Optins.ParentMessageId)
	// 如果返回404，说明会话不存在，重新获取会话
	if err != nil {
		if err.Error() == "send message failed: 404 Not Found" {
			stream, err = cli.GetChatStream(req.Prompt)
		}
	}
	if err != nil {
		resp := g.Map{
			"status":  "Error",
			"message": err.Error(),
		}
		r.Response.WriteJsonExit(resp)
	}
	// 流式回应
	res := &ChatProcessResponse{}
	rw := r.Response.RawWriter()
	flusher, ok := rw.(http.Flusher)
	if !ok {
		g.Log().Error(ctx, "rw.(http.Flusher) error")
		r.Response.WriteStatusExit(500)
		return
	}
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	for text := range stream.Stream {
		// g.DumpWithType(text)
		res.Id = text.MessageID
		res.Text = text.Content
		res.Role = "assistant"
		res.ConversationId = text.ConversationID
		res.ParentMessageId = req.Optins.ParentMessageId
		data := gjson.New(res).MustToJson()
		_, err = fmt.Fprintf(rw, "%s\n", data)
		if err != nil {
			g.Log().Error(ctx, "fmt.Fprintf error", err)
			break
		}
		flusher.Flush()
	}
	g.Log().Debug(ctx, "stream closed")
}
