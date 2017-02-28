package lib

import (
	"log"
	"os"
)

type Logger struct {
	Info  func(...interface{})
	Warn  func(...interface{})
	Error func(...interface{})
	Debug func(...interface{})
}

func NewLogger(debug bool) Logger {
	var (
		logger Logger
	)

	logger.Info = log.New(
		os.Stdout,
		"INFO:  ",
		log.Ldate|log.Ltime|log.Lshortfile,
	).Println

	logger.Warn = log.New(
		os.Stdout,
		"WARN:  ",
		log.Ldate|log.Ltime|log.Lshortfile,
	).Println

	logger.Error = log.New(
		os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	).Fatalln

	logger.Debug = log.New(
		os.Stdout,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	).Println

	return logger
}
