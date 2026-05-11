package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AboutHandler struct{}

func NewAboutHandler() *AboutHandler {
	return &AboutHandler{}
}

func (h *AboutHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", gin.H{
		"Title": "关于",
	})
}
