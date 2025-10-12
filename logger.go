package main

import (
	"fmt"
	"log"
	"os"
)

// Logger provides structured logging with different levels
type Logger struct {
	*log.Logger
	level LogLevel
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var logger *Logger

func init() {
	logger = NewLogger(INFO)
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		level:  level,
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("[DEBUG] "+format, args...)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.Printf("[INFO] "+format, args...)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.Printf("[WARN] "+format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.Printf("[ERROR] "+format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}

// SafeConfigRead provides better error handling for config file operations
func SafeConfigRead(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", configFile)
	}
	
	// Check if file is readable
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("cannot read configuration file: %v", err)
	}
	defer file.Close()
	
	return nil
}

// GracefulShutdown handles cleanup operations
type GracefulShutdown struct {
	shutdownFuncs []func() error
}

func NewGracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		shutdownFuncs: make([]func() error, 0),
	}
}

func (g *GracefulShutdown) Register(fn func() error) {
	g.shutdownFuncs = append(g.shutdownFuncs, fn)
}

func (g *GracefulShutdown) Shutdown() {
	logger.Info("Starting graceful shutdown...")
	for i := len(g.shutdownFuncs) - 1; i >= 0; i-- {
		if err := g.shutdownFuncs[i](); err != nil {
			logger.Error("Error during shutdown: %v", err)
		}
	}
	logger.Info("Shutdown completed")
}