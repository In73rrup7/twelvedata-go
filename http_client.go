// Portions of this file are adapted from github.com/spacecodewor/fmpcloud-go
// Copyright (c) 2021 Igor Churbakov
// Licensed under the MIT License -- see LICENSE file for details

package twelvedata

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type HTTPClient struct {
	logger        *zap.Logger
	client        *resty.Client
	apiKey        string
	retryCount    *int
	retryWaitTime *time.Duration
}

var (
	// apiCreditsLeft stores the last known API credits remaining from TwelveData response headers
	apiCreditsLeft int = -1
	// apiCreditsMu protects concurrent access to apiCreditsLeft
	apiCreditsMu sync.Mutex
)

func (h *HTTPClient) Get(endpoint string, data map[string]string) (response *resty.Response, err error) {
	if data == nil {
		data = make(map[string]string)
	}

	data["apikey"] = h.apiKey

	retries := 0
	for retries < *h.retryCount {
		response, err = h.client.R().
			SetQueryParams(data).
			Get(endpoint)

		if err != nil || response.StatusCode() != http.StatusOK {
			time.Sleep(*h.retryWaitTime)
			retries++

			// response is not valid when there is an error
			var errOrStatusCode zapcore.Field
			if err != nil {
				errOrStatusCode = zap.Error(err)
			} else {
				errOrStatusCode = zap.Int("statusCode", response.StatusCode())
			}

			h.logger.Info(
				"Retry request",
				zap.Int("retries", retries),
				errOrStatusCode,
				zap.String("endpoint", endpoint),
				zap.Any("data", data),
			)

			continue
		}

		// If we get here, the request was successful
		// Capture API credits from response headers
		if creditsLeftStr := response.Header().Get("api-credits-left"); creditsLeftStr != "" {
			if credits, parseErr := strconv.Atoi(creditsLeftStr); parseErr == nil {
				apiCreditsMu.Lock()
				apiCreditsLeft = credits
				apiCreditsMu.Unlock()
				h.logger.Debug(
					"API credits captured",
					zap.Int("creditsLeft", credits),
					zap.String("endpoint", endpoint),
				)
			}
		}
		break
	}

	if err == nil && response.StatusCode() != http.StatusOK {
		return response, fmt.Errorf("API returned non-200 status code: %d", response.StatusCode())
	}
	return response, err
}

// GetAPICreditsLeft returns the last known API credits remaining from TwelveData
// Returns -1 if no API call has been made yet
func GetAPICreditsLeft() int {
	apiCreditsMu.Lock()
	defer apiCreditsMu.Unlock()
	return apiCreditsLeft
}
