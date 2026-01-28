package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type HTTPResult struct {
	StatusCode int
	Body       string
}

type HTTPExecutor struct {
	client   *http.Client
	logLimit int
}

func NewHTTPExecutor(logLimit int) *HTTPExecutor {
	return &HTTPExecutor{
		client:   &http.Client{Timeout: 5 * time.Minute},
		logLimit: logLimit,
	}
}

func (e *HTTPExecutor) Run(ctx context.Context, job *domain.JobDefinition) (HTTPResult, error) {
	if job.Type != domain.JobTypeHTTP {
		return HTTPResult{}, fmt.Errorf("not an http job")
	}
	method := job.HTTPMethod
	if method == "" {
		method = http.MethodGet
	}
	timeout := time.Duration(job.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(runCtx, method, job.HTTPURL, strings.NewReader(job.HTTPBody))
	if err != nil {
		return HTTPResult{}, err
	}
	for k, v := range job.HTTPHeaders {
		req.Header.Set(k, v)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return HTTPResult{}, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(e.logLimit)))
	if err != nil {
		return HTTPResult{}, err
	}
	expected := job.ExpectedStatus
	if expected == 0 {
		expected = http.StatusOK
	}
	if resp.StatusCode != expected {
		return HTTPResult{StatusCode: resp.StatusCode, Body: string(bodyBytes)},
			fmt.Errorf("unexpected status %d want %d", resp.StatusCode, expected)
	}
	return HTTPResult{StatusCode: resp.StatusCode, Body: string(bodyBytes)}, nil
}
