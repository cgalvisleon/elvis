package middleware

import (
	"bytes"
	"context"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	lg "github.com/cgalvisleon/elvis/logs"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	// LogEntryCtxKey is the context.Context key to store the request log entry.
	LogEntryCtxKey = &contextKey{"LogEntry"}

	// DefaultLogger is called by the Logger middleware handler to log each request.
	// Its made a package-level variable so that it can be reconfigured for custom
	// logging configurations.
	DefaultLogger func(next http.Handler) http.Handler
)

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white. Logger prints a
// request ID if one is provided.
//
// Alternatively, look at https://github.com/goware/httplog for a more in-depth
// http logger with structured logging support.
//
// IMPORTANT NOTE: Logger should go before any other middleware that may change
// the response, such as `middleware.Recoverer`. Example:
//
// ```go
// r := chi.NewRouter()
// r.Use(middleware.Logger)        // <--<< Logger should come before Recoverer
// r.Use(middleware.Recoverer)
// r.Get("/", handler)
// ```
func Logger(next http.Handler) http.Handler {
	return DefaultLogger(next)
}

// RequestLogger returns a logger handler using a custom LogFormatter.
func RequestLogger(f LogFormatter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := f.NewLogEntry(r)
			ww := NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			reqID := utility.NewId()
			endPoint := r.URL.Path
			method := r.Method
			hostName, _ := os.Hostname()
			var mTotal uint64
			var mUsed uint64
			var mFree uint64
			memory, err := mem.VirtualMemory()
			if err != nil {
				mFree = 0
				mTotal = 0
				mUsed = 0
			}
			mTotal = memory.Total
			mUsed = memory.Used
			mFree = memory.Total - memory.Used
			pFree := float64(mFree) / float64(mTotal) * 100
			requests_host := CallRequests(hostName)
			requests_endpoint := CallRequests(endPoint)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			defer func() {
				telemetry := et.Json{
					"reqID":     reqID,
					"date_time": t1,
					"host_name": hostName,
					"request": et.Json{
						"method":   method,
						"endpoint": endPoint,
						"status":   ww.Status(),
						"bytes":    ww.BytesWritten(),
						"header":   ww.Header(),
						"scheme":   scheme,
						"host":     r.Host,
					},
					"since": et.Json{
						"value": time.Since(t1),
						"unity": "Milliseconds",
					},
					"memory": et.Json{
						"unity":        "MB",
						"total":        mTotal / 1024 / 1024,
						"used":         mUsed / 1024 / 1024,
						"free":         mFree / 1024 / 1024,
						"percent_free": math.Floor(pFree*100) / 100,
					},
					"request_host": et.Json{
						"host":   requests_host.Tag,
						"day":    requests_host.Day,
						"hour":   requests_host.Hour,
						"minute": requests_host.Minute,
						"second": requests_host.Seccond,
						"limit":  requests_host.Limit,
					},
					"requests_endpoint": et.Json{
						"endpoint": requests_endpoint.Tag,
						"day":      requests_endpoint.Day,
						"hour":     requests_endpoint.Hour,
						"minute":   requests_endpoint.Minute,
						"second":   requests_endpoint.Seccond,
						"limit":    requests_endpoint.Limit,
					},
				}

				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)

				go event.Action("telemetry", telemetry)

				if requests_host.Seccond > requests_host.Limit {
					go event.Action("requests/overflow", telemetry)
				}
			}()

			w.Header().Set("Reqid", reqID)
			next.ServeHTTP(ww, WithLogEntry(r, entry))
		}

		return http.HandlerFunc(fn)
	}
}

// LogFormatter initiates the beginning of a new LogEntry per request.
// See DefaultLogFormatter for an example implementation.
type LogFormatter interface {
	NewLogEntry(r *http.Request) LogEntry
}

// LogEntry records the final log when a request completes.
// See defaultLogEntry for an example implementation.
type LogEntry interface {
	Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{})
	Panic(v interface{}, stack []byte)
}

// GetLogEntry returns the in-context LogEntry for a request.
func GetLogEntry(r *http.Request) LogEntry {
	entry, _ := r.Context().Value(LogEntryCtxKey).(LogEntry)
	return entry
}

// WithLogEntry sets the in-context LogEntry for a request.
func WithLogEntry(r *http.Request, entry LogEntry) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LogEntryCtxKey, entry))
	return r
}

// LoggerInterface accepts printing to stdlib logger or compatible logger.
type LoggerInterface interface {
	Print(v ...interface{})
}

// DefaultLogFormatter is a simple logger that implements a LogFormatter.
type DefaultLogFormatter struct {
	Logger  LoggerInterface
	NoColor bool
}

// NewLogEntry creates a new LogEntry for the request.
func (l *DefaultLogFormatter) NewLogEntry(r *http.Request) LogEntry {
	entry := &defaultLogEntry{
		DefaultLogFormatter: l,
		request:             r,
		buf:                 &bytes.Buffer{},
	}

	reqID := GetReqID(r.Context())
	if reqID != "" {
		lg.CW(entry.buf, lg.NYellow, "[%s] ", reqID)
	}
	lg.CW(entry.buf, lg.NCyan, "")
	lg.CW(entry.buf, lg.BMagenta, "[%s]: ", r.Method)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	lg.CW(entry.buf, lg.NCyan, "%s://%s%s %s ", scheme, r.Host, r.RequestURI, r.Proto)

	entry.buf.WriteString("from ")
	entry.buf.WriteString(r.RemoteAddr)
	entry.buf.WriteString(" - ")

	return entry
}

type defaultLogEntry struct {
	*DefaultLogFormatter
	request *http.Request
	buf     *bytes.Buffer
}

func (l *defaultLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	switch {
	case status < 200:
		lg.CW(l.buf, lg.BBlue, "%03d", status)
	case status < 300:
		lg.CW(l.buf, lg.BGreen, "%03d", status)
	case status < 400:
		lg.CW(l.buf, lg.BCyan, "%03d", status)
	case status < 500:
		lg.CW(l.buf, lg.BYellow, "%03d", status)
	default:
		lg.CW(l.buf, lg.BRed, "%03d", status)
	}

	lg.CW(l.buf, lg.BBlue, " %dB", bytes)

	l.buf.WriteString(" in ")
	if elapsed < 500*time.Millisecond {
		lg.CW(l.buf, lg.NGreen, "%s", elapsed)
	} else if elapsed < 5*time.Second {
		lg.CW(l.buf, lg.NYellow, "%s", elapsed)
	} else {
		lg.CW(l.buf, lg.NRed, "%s", elapsed)
	}

	l.Logger.Print(l.buf.String())
}

func (l *defaultLogEntry) Panic(v interface{}, stack []byte) {
	PrintPrettyStack(v)
}

func init() {
	color := true
	if runtime.GOOS == "windows" {
		color = false
	}
	DefaultLogger = RequestLogger(&DefaultLogFormatter{Logger: log.New(os.Stdout, "", log.LstdFlags), NoColor: !color})
}
