package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.infoLog.Printf("%s %v", msg, fields)
	} else {
		l.infoLog.Println(msg)
	}
}

func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}
	if len(fields) > 0 {
		l.errorLog.Printf("%s %v", msg, fields)
	} else {
		l.errorLog.Println(msg)
	}
}

func (l *Logger) Fatal(msg string, err error) {
	l.errorLog.Fatalf("%s: %v", msg, err)
}

// Middleware для логирования HTTP запросов
func (l *Logger) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
		)

		next(w, r)

		l.Info("Response",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	}
}
