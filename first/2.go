package main

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

/*
	结合gin框架和zap日志库的使用,也可以使用第三方库实现的gin和zap的封装 https://github.com/gin-contrib/zap
*/

var zapLogger1 *zap.Logger

func main() {
	InitZapLogger1()
	engine := gin.New()
	engine.Use(ginZapLogger(zapLogger1), ginZapRecovery(zapLogger1, true))
	engine.GET("/1", func(c *gin.Context) {
		c.String(http.StatusOK, "你好,%s", "阿冰")
	})
	engine.GET("/2", func(c *gin.Context) {
		panic("1111")
	})
	engine.Run()
}

//普通日志
func ginZapLogger(logger *zap.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		//记录时间
		start := time.Now()
		//请求的url
		path := c.Request.URL.Path
		//请求的参数
		query := c.Request.URL.RawQuery
		//放行
		c.Next()
		//回到此处为解为时间
		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

func ginZapRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func InitZapLogger1() {
	//初始化logger配置
	//初始化编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	//使用第三方框架指定日志位置，并切割日志
	fileLogger := &lumberjack.Logger{
		Filename:   "./log/log_2.txt",
		MaxSize:    5,
		MaxAge:     10,
		MaxBackups: 30,
	}
	writeSyncer := zapcore.AddSync(fileLogger)
	//配置日志级别
	logLevel := zapcore.DebugLevel
	core := zapcore.NewCore(encoder, writeSyncer, logLevel)
	//创建logger,附加调用者信息
	zapLogger1 = zap.New(core, zap.AddCaller())
}
