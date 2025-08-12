package models

import (
	"time"
)

// BinanceBalance represents a spot wallet balance from Binance API
type BinanceBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// BinanceAccountInfo represents account information from Binance API
type BinanceAccountInfo struct {
	MakerCommission  int              `json:"makerCommission"`
	TakerCommission  int              `json:"takerCommission"`
	BuyerCommission  int              `json:"buyerCommission"`
	SellerCommission int              `json:"sellerCommission"`
	CanTrade         bool             `json:"canTrade"`
	CanWithdraw      bool             `json:"canWithdraw"`
	CanDeposit       bool             `json:"canDeposit"`
	UpdateTime       int64            `json:"updateTime"`
	AccountType      string           `json:"accountType"`
	Balances         []BinanceBalance `json:"balances"`
	Permissions      []string         `json:"permissions"`
}

// BinanceTicker represents 24hr ticker price change statistics
type BinanceTicker struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	Count              int    `json:"count"`
}

// BinanceKline represents candlestick/kline data
type BinanceKline struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	NumberOfTrades           int    `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
}

// BinancePriceData represents processed price information for a symbol
type BinancePriceData struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// BinancePortfolioData represents processed portfolio information
type BinancePortfolioData struct {
	AccountType string             `json:"account_type"`
	UpdateTime  time.Time          `json:"update_time"`
	Balances    map[string]float64 `json:"balances"` // asset -> quantity
	Prices      map[string]float64 `json:"prices"`   // symbol -> price
	Values      map[string]float64 `json:"values"`   // asset -> value in quote currency
	TotalValue  float64            `json:"total_value"`
	Permissions []string           `json:"permissions"`
}
