package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"products": []string{"Product A", "Product B"},
	})
}
