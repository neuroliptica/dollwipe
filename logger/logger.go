package logger

import (
	"fmt"
	"log"
)

var (
	MainLogger chan string

	Init, Cache, Info, Cookies, Proxies, Files, Captions Logger
)

func init() {
	MainLogger = make(chan string, 0)
	Init = MakeLogger("init", MainLogger)
	Cache = MakeLogger("cache", MainLogger)
	Info = MakeLogger("info", MainLogger)
	Cookies = MakeLogger("cookies", MainLogger)
	Proxies = MakeLogger("proxies", MainLogger)
	Files = MakeLogger("files", MainLogger)
	Captions = MakeLogger("captions", MainLogger)
}

// Logging everything through channel with messages queue.
func GlobalLogger(channel chan string) {
	for msg := range MainLogger {
		log.Println(msg)
	}
}

type Logger struct {
	LogType     string
	Destination chan string
}

func MakeLogger(logtype string, dest chan string) Logger {
	return Logger{logtype, dest}
}

func (l *Logger) Log(msg ...interface{}) {
	l.Destination <- fmt.Sprintf("[%s] %3s",
		l.LogType, fmt.Sprint(msg...))
}

func (l Logger) Logf(format string, msg ...interface{}) {
	l.Log(fmt.Sprintf(format, msg...))
}
