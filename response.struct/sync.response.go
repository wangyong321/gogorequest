package response_struct

import (
	req "gogorequest/request.struct"
	"net/http"
	"time"
)

type SyncResponse struct {
	Status      bool
	Error       error
	Request     *req.SyncRequestBody
	Response    *http.Response
	StatusCode  int
	Text        string
	StartTime   time.Time
	EndTime     time.Time
	ConsumeTime float64
}
