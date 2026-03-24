package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	logdomain "gpilot/internal/domain/log"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type lokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func (c *Client) Query(ctx context.Context, q logdomain.LogQuery) (*logdomain.LogQueryResult, error) {
	params := url.Values{}
	params.Set("query", q.Query)
	params.Set("limit", strconv.Itoa(q.Limit))
	params.Set("start", strconv.FormatInt(q.From.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(q.To.UnixNano(), 10))
	if q.Direction != "" {
		params.Set("direction", q.Direction)
	} else {
		params.Set("direction", "backward")
	}

	reqURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create loki request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read loki response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loki returned status %d: %s", resp.StatusCode, string(body))
	}

	var lokiResp lokiQueryResponse
	if err := json.Unmarshal(body, &lokiResp); err != nil {
		return nil, fmt.Errorf("parse loki response: %w", err)
	}

	var entries []logdomain.LogEntry
	for _, stream := range lokiResp.Data.Result {
		for _, val := range stream.Values {
			if len(val) < 2 {
				continue
			}
			ts, _ := strconv.ParseInt(val[0], 10, 64)
			entries = append(entries, logdomain.LogEntry{
				Timestamp: time.Unix(0, ts),
				Line:      val[1],
				Labels:    stream.Stream,
			})
		}
	}

	return &logdomain.LogQueryResult{
		Entries:   entries,
		Total:     len(entries),
		QueryExpr: q.Query,
	}, nil
}
