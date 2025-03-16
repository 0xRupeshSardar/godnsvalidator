package output

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/0xRupeshSardar/godnsvalidator/internal/config"
)

var (
	fileHandle *os.File
	fileMu     sync.Mutex
)

func Init(cfg *config.Config) {
	if cfg.NoColor {
		color.NoColor = true
	}

	if cfg.OutputFile != "" {
		f, err := os.Create(cfg.OutputFile)
		if err != nil {
			Error("Error creating output file: %v", err)
			return
		}
		fileHandle = f
	}
}

func Success(format string, args ...interface{}) {
	color.Green(format, args...)
}

func Error(format string, args ...interface{}) {
	color.Red(format, args...)
}

func LogServer(server, message string, cfg *config.Config) {
	if cfg.Silent {
		return
	}

	color.Cyan("[%s] %s: %s", 
		time.Now().Format("15:04:05"), 
		server, 
		message,
	)

	if fileHandle != nil {
		fileMu.Lock()
		defer fileMu.Unlock()
		fmt.Fprintf(fileHandle, "[%s] %s: %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			server,
			message,
		)
	}
}

func WriteResults(cfg *config.Config) {
    if fileHandle != nil {
        fileMu.Lock()
        defer fileMu.Unlock()

        if err := fileHandle.Sync(); err != nil {
            Error("Error flushing file: %v", err)
        }

        if err := fileHandle.Close(); err != nil {
            Error("Error closing file: %v", err)
        }
    }
}