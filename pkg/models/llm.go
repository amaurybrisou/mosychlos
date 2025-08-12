package models

import (
	"fmt"
	"time"
)

type RateLimitConfig struct {
	Enabled      bool          `yaml:"enabled"`
	BaseDelay    time.Duration `yaml:"base_delay"` // min sleep
	MaxDelay     time.Duration `yaml:"max_delay"`  // cap
	JitterFactor float64       `yaml:"jitter_factor"`
	LogMetrics   bool          `yaml:"log_metrics"`
}

// Validate validates the RateLimitConfig
func (rlc *RateLimitConfig) Validate() error {
	// BaseDelay must be non-negative
	if rlc.BaseDelay < 0 {
		return fmt.Errorf("BaseDelay must be non-negative, got: %v", rlc.BaseDelay)
	}

	// MaxDelay must be non-negative
	if rlc.MaxDelay < 0 {
		return fmt.Errorf("MaxDelay must be non-negative, got: %v", rlc.MaxDelay)
	}

	// MaxDelay must be >= BaseDelay if both are positive
	if rlc.BaseDelay > 0 && rlc.MaxDelay > 0 && rlc.MaxDelay < rlc.BaseDelay {
		return fmt.Errorf("MaxDelay (%v) must be greater than or equal to BaseDelay (%v)", rlc.MaxDelay, rlc.BaseDelay)
	}

	// JitterFactor must be between 0 and 1
	if rlc.JitterFactor < 0 || rlc.JitterFactor > 1 {
		return fmt.Errorf("JitterFactor must be between 0 and 1, got: %f", rlc.JitterFactor)
	}

	return nil
}

type RetryConfig struct {
	MaxRetries      int           `yaml:"max_retries"`
	BaseDelay       time.Duration `yaml:"base_delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	ExponentialBase float64       `yaml:"exponential_base"`
	JitterFactor    float64       `yaml:"jitter_factor"`
}

// Validate validates the RetryConfig
func (rc *RetryConfig) Validate() error {
	// MaxRetries must be non-negative
	if rc.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be non-negative, got: %d", rc.MaxRetries)
	}

	// BaseDelay must be non-negative
	if rc.BaseDelay < 0 {
		return fmt.Errorf("BaseDelay must be non-negative, got: %v", rc.BaseDelay)
	}

	// MaxDelay must be non-negative
	if rc.MaxDelay < 0 {
		return fmt.Errorf("MaxDelay must be non-negative, got: %v", rc.MaxDelay)
	}

	// MaxDelay must be >= BaseDelay if both are positive
	if rc.BaseDelay > 0 && rc.MaxDelay > 0 && rc.MaxDelay < rc.BaseDelay {
		return fmt.Errorf("MaxDelay (%v) must be greater than or equal to BaseDelay (%v)", rc.MaxDelay, rc.BaseDelay)
	}

	// ExponentialBase must be greater than 1
	if rc.ExponentialBase <= 1 {
		return fmt.Errorf("ExponentialBase must be greater than 1, got: %f", rc.ExponentialBase)
	}

	// JitterFactor must be between 0 and 1
	if rc.JitterFactor < 0 || rc.JitterFactor > 1 {
		return fmt.Errorf("JitterFactor must be between 0 and 1, got: %f", rc.JitterFactor)
	}

	return nil
}
