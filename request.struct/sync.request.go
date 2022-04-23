package request_struct

import (
	"gogorequest/engine"
	"time"
)

type SyncRequestBody struct {
	URL         string
	Headers     map[string]string
	Method      string
	Body        interface{}
	Spider      *engine.SyncEngine
	Proxy       string
	Timeout     time.Duration
	Meta        map[string]interface{}
	RetryNumber int64
	StartTime   time.Time
}
