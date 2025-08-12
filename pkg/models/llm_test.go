package models

import (
	"testing"
	"time"
)

func TestRateLimitConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		config  RateLimitConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RateLimitConfig{
				Enabled:      true,
				BaseDelay:    1 * time.Second,
				MaxDelay:     10 * time.Second,
				JitterFactor: 0.5,
				LogMetrics:   true,
			},
			wantErr: false,
		},
		{
			name: "negative base delay",
			config: RateLimitConfig{
				BaseDelay:    -1 * time.Second,
				MaxDelay:     10 * time.Second,
				JitterFactor: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative max delay",
			config: RateLimitConfig{
				BaseDelay:    1 * time.Second,
				MaxDelay:     -1 * time.Second,
				JitterFactor: 0.5,
			},
			wantErr: true,
		},
		{
			name: "max delay less than base delay",
			config: RateLimitConfig{
				BaseDelay:    10 * time.Second,
				MaxDelay:     5 * time.Second,
				JitterFactor: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid jitter factor high",
			config: RateLimitConfig{
				BaseDelay:    1 * time.Second,
				MaxDelay:     10 * time.Second,
				JitterFactor: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid jitter factor low",
			config: RateLimitConfig{
				BaseDelay:    1 * time.Second,
				MaxDelay:     10 * time.Second,
				JitterFactor: -0.1,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.config.Validate()

			if c.wantErr && err == nil {
				t.Error("expected error but got nil")
			}

			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRetryConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		config  RetryConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RetryConfig{
				MaxRetries:      3,
				BaseDelay:       1 * time.Second,
				MaxDelay:        10 * time.Second,
				ExponentialBase: 2.0,
				JitterFactor:    0.1,
			},
			wantErr: false,
		},
		{
			name: "negative max retries",
			config: RetryConfig{
				MaxRetries:      -1,
				BaseDelay:       1 * time.Second,
				MaxDelay:        10 * time.Second,
				ExponentialBase: 2.0,
				JitterFactor:    0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid exponential base",
			config: RetryConfig{
				MaxRetries:      3,
				BaseDelay:       1 * time.Second,
				MaxDelay:        10 * time.Second,
				ExponentialBase: 1.0,
				JitterFactor:    0.1,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.config.Validate()

			if c.wantErr && err == nil {
				t.Error("expected error but got nil")
			}

			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestToolsRateLimit_Validate(t *testing.T) {
	cases := []struct {
		name    string
		config  ToolsRateLimit
		wantErr bool
	}{
		{
			name: "valid config",
			config: ToolsRateLimit{
				RequestsPerSecond: 10,
				RequestsPerDay:    1000,
				Burst:             20,
			},
			wantErr: false,
		},
		{
			name: "negative requests per second",
			config: ToolsRateLimit{
				RequestsPerSecond: -1,
				RequestsPerDay:    1000,
				Burst:             20,
			},
			wantErr: true,
		},
		{
			name: "negative requests per day",
			config: ToolsRateLimit{
				RequestsPerSecond: 10,
				RequestsPerDay:    -1,
				Burst:             20,
			},
			wantErr: true,
		},
		{
			name: "negative burst",
			config: ToolsRateLimit{
				RequestsPerSecond: 10,
				RequestsPerDay:    1000,
				Burst:             -1,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.config.Validate()

			if c.wantErr && err == nil {
				t.Error("expected error but got nil")
			}

			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
