package xyhelper

import "github.com/gogf/gf/v2/frame/g"

func init() {
	s := g.Server()
	group := s.Group("/api")
	group.POST("/session", Session)
	group.POST("/chat-process", ChatProcess)

}
