package app

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"goquizbox/internal/buildinfo"
	"goquizbox/internal/logger"
	"goquizbox/internal/repos"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleHealthz() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		hostName, err := os.Hostname()
		if err != nil {
			logger.Errorf("could not get hostname: %v", err)
		}

		conn, err := s.env.Database().Pool.Acquire(ctx)
		if err != nil {
			logger.Errorf("failed to acquire database connection: 5v", err)
		}
		defer conn.Release()

		dbUp := true
		if err := conn.Conn().Ping(ctx); err != nil {
			dbUp = false
			logger.Errorf("failed to ping database: %v", err)
		}

		checkDB := repos.NewCheckerDB(s.env.Database())
		hello, err := checkDB.SelectOne(ctx)
		if err != nil {
			logger.Errorf("could not db select 1: %v", err)
		}

		version, err := checkDB.SelectVersion(ctx)
		if err != nil {
			logger.Errorf("could not db version: %v", err)
		}

		memStats := &runtime.MemStats{}
		runtime.ReadMemStats(memStats)

		currentTime := time.Now()
		tZone, offset := currentTime.Zone()

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"env":        os.Getenv("ENV"),
			"build_time": "1",
			"build_id":   buildinfo.BuildID,
			"build_tag":  buildinfo.BuildTag,
			"time": map[string]interface{}{
				"now":      currentTime,
				"timezone": tZone,
				"offset":   offset,
			},
			"db": map[string]interface{}{
				"type":    "postgres",
				"up":      dbUp,
				"hello":   hello,
				"version": version,
			},
			"server": map[string]interface{}{
				"hostname":   hostName,
				"cpu":        runtime.NumCPU(),
				"goroutines": runtime.NumGoroutine(),
				"goarch":     runtime.GOARCH,
				"goos":       runtime.GOOS,
				"compiler":   runtime.Compiler,
				"memory": map[string]interface{}{
					"alloc":       fmt.Sprintf("%v MB", bytesToMb(memStats.Alloc)),
					"total_alloc": fmt.Sprintf("%v MB", bytesToMb(memStats.TotalAlloc)),
					"sys":         fmt.Sprintf("%v MB", bytesToMb(memStats.Sys)),
					"num_gc":      memStats.NumGC,
				},
			},
		})
	}
}

func bytesToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
