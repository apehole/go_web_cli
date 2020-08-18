/**
 *@Author: IronHuang
 *@Date: 2020/8/18 9:24 下午
**/

package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go_web_cli/settings"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// 初始化logger
func Init(cfg *settings.LogConfig) (err error) {
	writeSyncer := getLogWriter(cfg.FileName,
		cfg.MaxSize,
		cfg.MaxBackups,
		cfg.MaxAge)
	encoder := getEncoder()
	// zapcore.DebugLevel 打印级别
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writeSyncer, l)
	// zap.AddCaller 返回调用位置
	lg := zap.New(core, zap.AddCaller())
	// 替换zap库中全局的logger
	zap.ReplaceGlobals(lg)
	return
}

func getLogWriter(fileName string, maxSize, maxBackups, maxAge int) zapcore.WriteSyncer {
	//file, _ := os.OpenFile("./test.log",os.O_APPEND|os.O_RDWR|os.O_CREATE,0744)
	lumberJackLogger := &lumberjack.Logger{
		Filename: fileName,
		// 限制大小M
		MaxSize: maxSize,
		// 备份数量
		MaxBackups: maxBackups,
		// 最大保存天数
		MaxAge: maxAge,
		// 是否压缩
		//Compress: false,
	}
	//return zapcore.AddSync(file)
	return zapcore.AddSync(lumberJackLogger)
}

func myEncodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// 自定义时间格式
func myEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	myEncodeTimeLayout(t, "2006-01-02 15:04:05", enc)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = myEncodeTime
	return zapcore.NewJSONEncoder(encoderConfig)
}

// 定义GinLogger中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		zap.L().Info(path,
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

// 定义Recover中间件,stack 是否显示栈信息
func GinRecover(stack bool) gin.HandlerFunc {
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
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
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
