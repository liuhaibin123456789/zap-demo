package main

import (
	"net/http"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*
	自定义zap配置
*/

var sugarLogger *zap.SugaredLogger

func main() {
	InitZapLogger()
	for i := 0; i < 100000; i++ {
		sugarLogger.Infof("test...%d", i)
	}
	simpleHttpGet1("www.sogo.com")
	simpleHttpGet1("http://www.baidu.com")
}

func InitZapLogger() {
	//指定编码器，使用默认编码器
	productionEncoderConfig := zap.NewProductionEncoderConfig()
	//修改时间编码
	productionEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//修改大写日志级别
	productionEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//调用方信息

	encoder := zapcore.NewConsoleEncoder(productionEncoderConfig)
	//file, err := os.Create("./log.txt")
	//if err != nil {
	//	return
	//}
	//指定日志文件位置,使用第三方库对日志进行切割
	file := &lumberjack.Logger{
		Filename:   "./log/log.txt",
		MaxSize:    1,
		MaxAge:     5,
		MaxBackups: 30,
		//是否压缩文件，默认否
		Compress: false,
	}
	writer := zapcore.AddSync(file)
	//指定日志级别
	levelEnabler := zapcore.DebugLevel
	core := zapcore.NewCore(encoder, writer, levelEnabler)
	//增加调用方信息
	sugarLogger = zap.New(core, zap.AddCaller()).Sugar()
	return
}

func simpleHttpGet1(url string) {
	sugarLogger.Debugf("Trying to hit GET request for %s", url)
	resp, err := http.Get(url)
	if err != nil {
		sugarLogger.Errorf("Error fetching URL %s : Error = %s", url, err)
	} else {
		sugarLogger.Infof("Success! statusCode = %s for URL %s", resp.Status, url)
		resp.Body.Close()
	}
}
