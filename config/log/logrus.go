package log

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"refina-transaction/config/env"
	"refina-transaction/internal/utils/data"

	"github.com/sirupsen/logrus"
)

// ApacheStyleFormatter - custom formatter dengan Apache/Nginx style dan color support
type ApacheStyleFormatter struct {
	NoColors bool
}

// Format - implementasi interface Formatter
func (f *ApacheStyleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// Color codes untuk setiap level
	var levelColor string
	var resetColor string = "\x1b[0m"

	if !f.NoColors {
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = "\x1b[36m" // Cyan
		case logrus.InfoLevel:
			levelColor = "\x1b[32m" // Green
		case logrus.WarnLevel:
			levelColor = "\x1b[33m" // Yellow
		case logrus.ErrorLevel:
			levelColor = "\x1b[31m" // Red
		case logrus.FatalLevel, logrus.PanicLevel:
			levelColor = "\x1b[35m" // Magenta
		case logrus.TraceLevel:
			levelColor = "\x1b[37m" // White
		default:
			levelColor = "\x1b[0m" // Default
		}
	}

	// Format timestamp dalam style Apache: [18/Aug/2024:14:30:25 +0700]
	timestamp := entry.Time.Format("02/Jan/2006:15:04:05 -0700")

	// Level string dengan padding
	level := strings.ToUpper(entry.Level.String())

	// Format dasar: [timestamp] LEVEL: message
	if !f.NoColors {
		fmt.Fprintf(b, "[%s] %s%s%s: %s",
			timestamp,
			levelColor,
			level,
			resetColor,
			entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] %s: %s",
			timestamp,
			level,
			entry.Message)
	}

	// Tambahkan fields jika ada
	if len(entry.Data) > 0 {
		fmt.Fprintf(b, " - ")

		fieldCount := 0
		for key, value := range entry.Data {
			if fieldCount > 0 {
				fmt.Fprintf(b, ", ")
			}

			// Format value berdasarkan tipe
			var valueStr string
			switch v := value.(type) {
			case string:
				// Quote string values yang mengandung spasi atau karakter khusus
				if strings.Contains(v, " ") || strings.Contains(v, ",") || strings.Contains(v, "=") {
					valueStr = fmt.Sprintf(`"%s"`, v)
				} else {
					valueStr = v
				}
			default:
				valueStr = fmt.Sprintf("%v", v)
			}

			fmt.Fprintf(b, "%s: %s", key, valueStr)
			fieldCount++
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

var Log *logrus.Logger

func SetupLogger() {
	Log = logrus.New()

	switch env.Cfg.Server.Mode {
	case data.PRODUCTION_MODE:
		Log.SetLevel(logrus.InfoLevel)

		// Gunakan JSON formatter untuk production
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})

		file, err := os.OpenFile(".server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			Log.Fatal("Failed to open log file:", err)
		}
		Log.SetOutput(file)
		Log.Debug("Production Log")

	case data.STAGING_MODE:
		Log.SetLevel(logrus.TraceLevel)

		// Gunakan Apache style formatter tanpa colors untuk staging (file output)
		Log.SetFormatter(&ApacheStyleFormatter{
			NoColors: true,
		})

		Log.SetOutput(os.Stdout)
		Log.Debug("Staging Log")

	default: // Development mode
		Log.SetLevel(logrus.TraceLevel)

		// Gunakan Apache style formatter dengan colors untuk development (console output)
		Log.SetFormatter(&ApacheStyleFormatter{
			NoColors: false,
		})

		Log.SetOutput(os.Stdout)
		Log.Debug("Development Log")
	}

	Log.Debug(env.Cfg.Server.Mode)
}

// Helper functions untuk logging yang lebih mudah digunakan
func Info(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Info(msg)
}

func Error(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Error(msg)
}

func Warn(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Warn(msg)
}

func Debug(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Debug(msg)
}

func Fatal(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Fatal(msg)
}

func Trace(msg string, fields ...map[string]any) {
	entry := Log.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = Log.WithFields(logrus.Fields(fields[0]))
	}
	entry.Trace(msg)
}
