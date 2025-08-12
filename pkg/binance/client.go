package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// client implements the Client interface for Binance REST API
type client struct {
	config     *config.BinanceConfig
	httpClient *http.Client
	baseURL    string
}

// New creates a new Binance client with the given configuration
func New(cfg *config.BinanceConfig) Client {
	baseURL := "https://api.binance.com"
	if cfg.BaseURL != "" {
		baseURL = strings.TrimRight(cfg.BaseURL, "/")
	}

	return &client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

// GetAccountInfo retrieves spot account information including balances
func (c *client) GetAccountInfo(ctx context.Context) (*models.BinanceAccountInfo, error) {
	endpoint := "/api/v3/account"

	resp, err := c.makeAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var accountInfo models.BinanceAccountInfo
	if err := json.NewDecoder(resp.Body).Decode(&accountInfo); err != nil {
		return nil, fmt.Errorf("failed to decode account info: %w", err)
	}

	return &accountInfo, nil
}

// GetTicker24hr gets 24hr ticker price change statistics for a symbol
func (c *client) GetTicker24hr(ctx context.Context, symbol string) (*models.BinanceTicker, error) {
	endpoint := "/api/v3/ticker/24hr"
	params := url.Values{"symbol": {strings.ToUpper(symbol)}}

	resp, err := c.makePublicRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var ticker models.BinanceTicker
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return nil, fmt.Errorf("failed to decode ticker: %w", err)
	}

	return &ticker, nil
}

// GetAllTickers24hr gets 24hr ticker price change statistics for all symbols
func (c *client) GetAllTickers24hr(ctx context.Context) ([]*models.BinanceTicker, error) {
	endpoint := "/api/v3/ticker/24hr"

	resp, err := c.makePublicRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tickers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var tickers []models.BinanceTicker
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, fmt.Errorf("failed to decode tickers: %w", err)
	}

	// convert to pointer slice
	result := make([]*models.BinanceTicker, len(tickers))
	for i := range tickers {
		result[i] = &tickers[i]
	}

	return result, nil
}

// GetKlines gets candlestick/kline data for a symbol
func (c *client) GetKlines(ctx context.Context, symbol string, interval string, limit int, startTime, endTime *time.Time) ([]*models.BinanceKline, error) {
	endpoint := "/api/v3/klines"
	params := url.Values{
		"symbol":   {strings.ToUpper(symbol)},
		"interval": {interval},
		"limit":    {strconv.Itoa(limit)},
	}

	if startTime != nil {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if endTime != nil {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}

	resp, err := c.makePublicRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var rawKlines [][]any
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, fmt.Errorf("failed to decode klines: %w", err)
	}

	klines := make([]*models.BinanceKline, len(rawKlines))
	for i, raw := range rawKlines {
		if len(raw) < 12 {
			return nil, fmt.Errorf("invalid kline data format")
		}

		kline := &models.BinanceKline{
			OpenTime:                 int64(raw[0].(float64)),
			Open:                     raw[1].(string),
			High:                     raw[2].(string),
			Low:                      raw[3].(string),
			Close:                    raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                int64(raw[6].(float64)),
			QuoteAssetVolume:         raw[7].(string),
			NumberOfTrades:           int(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
		}
		klines[i] = kline
	}

	return klines, nil
}

// GetPrice gets current price for a symbol
func (c *client) GetPrice(ctx context.Context, symbol string) (*models.BinancePriceData, error) {
	endpoint := "/api/v3/ticker/price"
	params := url.Values{"symbol": {strings.ToUpper(symbol)}}

	resp, err := c.makePublicRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var priceResp struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return nil, fmt.Errorf("failed to decode price: %w", err)
	}

	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}

	return &models.BinancePriceData{
		Symbol:    priceResp.Symbol,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}

// GetPrices gets current prices for multiple symbols
func (c *client) GetPrices(ctx context.Context, symbols []string) ([]*models.BinancePriceData, error) {
	if len(symbols) == 0 {
		return []*models.BinancePriceData{}, nil
	}

	endpoint := "/api/v3/ticker/price"

	resp, err := c.makePublicRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var priceResponses []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResponses); err != nil {
		return nil, fmt.Errorf("failed to decode prices: %w", err)
	}

	// create a map for quick lookup
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		symbolMap[strings.ToUpper(symbol)] = true
	}

	var result []*models.BinancePriceData
	timestamp := time.Now()

	for _, priceResp := range priceResponses {
		if !symbolMap[priceResp.Symbol] {
			continue
		}

		price, err := strconv.ParseFloat(priceResp.Price, 64)
		if err != nil {
			continue // skip invalid prices
		}

		result = append(result, &models.BinancePriceData{
			Symbol:    priceResp.Symbol,
			Price:     price,
			Timestamp: timestamp,
		})
	}

	return result, nil
}

// makePublicRequest makes a request to a public API endpoint
func (c *client) makePublicRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	u := c.baseURL + endpoint
	if params != nil {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.httpClient.Do(req)
}

// makeAuthenticatedRequest makes a request to an authenticated API endpoint
func (c *client) makeAuthenticatedRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	if c.config.APIKey == "" || c.config.APISecret == "" {
		return nil, fmt.Errorf("API key and secret are required for authenticated requests")
	}

	// add timestamp and signature for authenticated requests
	if params == nil {
		params = url.Values{}
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// create signature (simplified - in production use proper HMAC-SHA256)
	signature := c.createSignature(params.Encode())
	params.Set("signature", signature)

	u := c.baseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	return c.httpClient.Do(req)
}

// createSignature creates HMAC SHA256 signature for authenticated requests
func (c *client) createSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.config.APISecret))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}
