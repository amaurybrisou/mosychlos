package llm

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/amaurybrisou/mosychlos/internal/config"

	bagmocks "github.com/amaurybrisou/mosychlos/pkg/bag/mocks"
	fsmocks "github.com/amaurybrisou/mosychlos/pkg/fs/mocks"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestNewBatchServiceFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockBag := bagmocks.NewMockSharedBag(ctrl)

	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "openai",
		},
	}

	factory := NewBatchServiceFactory(cfg, mockFS, mockBag)

	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)
	assert.Equal(t, mockFS, factory.filesystem)
	assert.Equal(t, mockBag, factory.sharedBag)
}

func TestBatchServiceFactory_CreateManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful creation with openai provider",
			config: &config.Config{
				DataDir:   "/tmp/data",
				CacheDir:  "/tmp/cache",
				ConfigDir: "/tmp/config",
				Localization: models.LocalizationConfig{
					Country:  "US",
					Language: "en",
					Currency: "USD",
					Timezone: "America/New_York",
					Region:   "NY",
					City:     "New York",
				},
				LLM: config.LLMConfig{
					Provider: "openai",
					Model:    config.LLMModelGPT4o,
					APIKey:   "test-api-key",
				},
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			config: &config.Config{
				LLM: config.LLMConfig{
					Provider: "",
				},
			},
			wantErr: true,
			errMsg:  "LLM provider not configured",
		},
		{
			name: "unsupported provider",
			config: &config.Config{
				LLM: config.LLMConfig{
					Provider: "claude",
				},
			},
			wantErr: true,
			errMsg:  "unsupported AI provider for batch processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := fsmocks.NewMockFS(ctrl)
			mockBag := bagmocks.NewMockSharedBag(ctrl)

			// Mock the bag operations needed for prompt manager creation
			mockBag.EXPECT().Snapshot().Return(nil).AnyTimes()

			factory := NewBatchServiceFactory(tt.config, mockFS, mockBag)

			manager, err := factory.CreateManager()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manager)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestBatchServiceFactory_createBatchClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		provider string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "openai provider",
			provider: "openai",
			wantErr:  false,
		},
		{
			name:     "empty provider",
			provider: "",
			wantErr:  true,
			errMsg:   "LLM provider not configured",
		},
		{
			name:     "unsupported provider",
			provider: "claude",
			wantErr:  true,
			errMsg:   "unsupported AI provider for batch processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := fsmocks.NewMockFS(ctrl)
			mockBag := bagmocks.NewMockSharedBag(ctrl)

			cfg := &config.Config{
				LLM: config.LLMConfig{
					Provider: tt.provider,
					APIKey:   "test-api-key",
				},
			}

			factory := NewBatchServiceFactory(cfg, mockFS, mockBag)

			client, err := factory.createBatchClient()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}
