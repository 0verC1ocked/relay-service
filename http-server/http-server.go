package httpserver

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartServer(addr string, rec chan []byte) {
	router := gin.Default()

	router.POST("/relayCreateMatch", func(ctx *gin.Context) {
		b, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Println("Error reading body", err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Error reading request data",
			})
			return
		}
		rec <- b
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Match Created",
		})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	s := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Unable to start server", err.Error())
	}
}
