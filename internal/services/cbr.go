package services

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"bankapi/internal/config"
)

type CBRService struct {
	endpoint     string
	fallbackRate float64
	cache        *redis.Client
	logger       *logrus.Logger
}

func NewCBRService(cfg config.Config, cache *redis.Client, logger *logrus.Logger) *CBRService {
	return &CBRService{endpoint: cfg.CBREndpoint, fallbackRate: cfg.CBRFallbackKeyRate, cache: cache, logger: logger}
}

func (s *CBRService) KeyRate(ctx context.Context) float64 {
	const cacheKey = "cbr:key_rate"
	if s.cache != nil {
		value, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			if parsed, parseErr := strconv.ParseFloat(value, 64); parseErr == nil {
				return parsed
			}
		}
	}

	rate, err := s.fetchKeyRate(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("using fallback key rate")
		return s.fallbackRate
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, fmt.Sprintf("%.4f", rate), 12*time.Hour).Err()
	}
	return rate
}

func (s *CBRService) fetchKeyRate(ctx context.Context) (float64, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -14).Format("2006-01-02")
	to := now.Format("2006-01-02")
	body := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>%s</fromDate>
      <ToDate>%s</ToDate>
    </KeyRate>
  </soap:Body>
</soap:Envelope>`, from, to)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBufferString(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("cbr status %d", resp.StatusCode)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return parseLastRate(payload)
}

func parseLastRate(payload []byte) (float64, error) {
	decoder := xml.NewDecoder(bytes.NewReader(payload))
	var latest string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || !strings.EqualFold(start.Name.Local, "Rate") {
			continue
		}
		var value string
		if err := decoder.DecodeElement(&value, &start); err != nil {
			return 0, err
		}
		latest = strings.ReplaceAll(strings.TrimSpace(value), ",", ".")
	}
	if latest == "" {
		return 0, fmt.Errorf("rate not found")
	}
	return strconv.ParseFloat(latest, 64)
}
