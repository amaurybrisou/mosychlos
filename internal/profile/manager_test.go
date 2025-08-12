package profile

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	pkgfs "github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// testFS wraps fstest.MapFS to implement the custom fs.FS interface
type testFS struct {
	fstest.MapFS
	dirs map[string]bool
}

func newTestFS() *testFS {
	return &testFS{
		MapFS: make(fstest.MapFS),
		dirs:  make(map[string]bool),
	}
}

func (tfs *testFS) ReadFile(path string) ([]byte, error) {
	file, err := tfs.MapFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, info.Size())
	_, err = file.Read(data)
	return data, err
}

func (tfs *testFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := tfs.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tfs.MapFS[path] = &fstest.MapFile{
		Data: data,
		Mode: perm,
	}
	return nil
}

func (tfs *testFS) MkdirAll(path string, perm fs.FileMode) error {
	// Mark directory as existing
	tfs.dirs[path] = true

	// Also mark all parent directories
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		tfs.MkdirAll(parent, perm)
	}
	return nil
}

func (tfs *testFS) Stat(path string) (fs.FileInfo, error) {
	if tfs.dirs[path] {
		// Return a fake directory info
		return &testFileInfo{name: filepath.Base(path), isDir: true}, nil
	}

	file, err := tfs.MapFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}

func (tfs *testFS) Remove(path string) error {
	delete(tfs.MapFS, path)
	delete(tfs.dirs, path)
	return nil
}

func (tfs *testFS) Rename(oldPath, newPath string) error {
	if file, exists := tfs.MapFS[oldPath]; exists {
		tfs.MapFS[newPath] = file
		delete(tfs.MapFS, oldPath)
	}
	return nil
}

type testFileInfo struct {
	name  string
	isDir bool
}

func (tfi *testFileInfo) Name() string       { return tfi.name }
func (tfi *testFileInfo) Size() int64        { return 0 }
func (tfi *testFileInfo) Mode() fs.FileMode  { return 0644 }
func (tfi *testFileInfo) ModTime() time.Time { return time.Now() }
func (tfi *testFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi *testFileInfo) Sys() any           { return nil }

func TestNewProfileManager(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		fs        pkgfs.FS
		dataDir   string
		sharedBag bag.SharedBag
		wantErr   bool
	}{
		{
			name:      "valid config",
			fs:        newTestFS(),
			dataDir:   "/tmp/data",
			sharedBag: bag.NewSharedBag(),
			wantErr:   false,
		},
		{
			name:      "nil filesystem",
			fs:        nil,
			dataDir:   "/tmp/data",
			sharedBag: bag.NewSharedBag(),
			wantErr:   true,
		},
		{
			name:      "empty dataDir",
			fs:        newTestFS(),
			dataDir:   "",
			sharedBag: bag.NewSharedBag(),
			wantErr:   true,
		},
		{
			name:      "nil shared bag is ok",
			fs:        newTestFS(),
			dataDir:   "/tmp/data",
			sharedBag: nil,
			wantErr:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			manager, err := NewProfileManager(c.fs, c.dataDir, c.sharedBag)

			if c.wantErr {
				require.Error(t, err)
				assert.Nil(t, manager)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, manager)
				assert.Equal(t, c.dataDir, manager.configDir)
			}
		})
	}
}

func TestProfileManager_LoadProfile(t *testing.T) {
	t.Parallel()

	// Create test filesystem with real config content
	testFS := newTestFS()
	testFS.MapFS["investment_profiles/defaults/FR/conservative.yaml"] = &fstest.MapFile{
		Data: []byte(`investment_style: 'income'
research_depth: 'basic'
risk_tolerance: 'conservative'

preferred_assets:
  - 'bonds'
  - 'dividend_equities'
  - 'real_estate'

avoided_sectors:
  - 'tobacco'
  - 'weapons'
  - 'fossil_fuels'

esg_criteria:
  esg_importance: 'important'
  esg_focus:
    - 'environmental'
    - 'social'
    - 'governance'
  exclusion_criteria:
    - 'fossil_fuels'
    - 'tobacco'
    - 'weapons'

regional_context:
  country: 'FR'
  language: 'fr'
  currency: 'EUR'
  timezone: 'Europe/Paris'
  preferred_asset_classes:
    - 'European equities'
    - 'Euro bonds'
  esg_preferences:
    environmental:
      - 'green energy'

profile_version: '1.0'
source: 'default_regional'
`),
	}

	testFS.MapFS["investment_profiles/defaults/global/moderate.yaml"] = &fstest.MapFile{
		Data: []byte(`investment_style: 'balanced'
research_depth: 'basic'
risk_tolerance: 'moderate'

preferred_assets:
  - 'equities'
  - 'bonds'

avoided_sectors: []

esg_criteria:
  esg_importance: 'low'
  esg_focus: []
  exclusion_criteria: []

regional_preferences:
  country: ''
  language: 'en'
  currency: 'USD'
  timezone: 'UTC'

profile_version: '1.0'
source: 'default_global'
`),
	}

	// Also add global conservative for fallback test
	testFS.MapFS["investment_profiles/defaults/global/conservative.yaml"] = &fstest.MapFile{
		Data: []byte(`investment_style: 'income'
research_depth: 'basic'
risk_tolerance: 'conservative'

preferred_assets:
  - 'bonds'
  - 'dividend_equities'

avoided_sectors: []

esg_criteria:
  esg_importance: 'low'
  esg_focus: []
  exclusion_criteria: []

regional_context:
  country: ''
  language: 'en'
  currency: 'USD'
  timezone: 'UTC'

profile_version: '1.0'
source: 'default_global'
`),
	}

	cases := []struct {
		name          string
		country       string
		riskTolerance string
		setupBag      func(bag.SharedBag)
		wantErr       bool
		expectSource  string // "country", "global", "bag"
	}{
		{
			name:          "load country specific profile",
			country:       "FR",
			riskTolerance: "conservative",
			setupBag:      func(bag.SharedBag) {}, // empty bag
			wantErr:       false,
			expectSource:  "country",
		},
		{
			name:          "fallback to global profile",
			country:       "US",
			riskTolerance: "moderate",
			setupBag:      func(bag.SharedBag) {}, // empty bag
			wantErr:       false,
			expectSource:  "global",
		},
		{
			name:          "return from shared bag",
			country:       "FR", // Use FR which has a file, but bag should be checked first
			riskTolerance: "conservative",
			setupBag: func(sb bag.SharedBag) {
				profile := &models.InvestmentProfile{
					InvestmentStyle: "growth",
					ResearchDepth:   "comprehensive",
				}
				sb.Set(keys.KProfile, profile)
			},
			wantErr:      false,
			expectSource: "bag",
		},
		{
			name:          "profile not found",
			country:       "XX",
			riskTolerance: "unknown",
			setupBag:      func(bag.SharedBag) {}, // empty bag
			wantErr:       true,
		},
		{
			name:          "empty parameters",
			country:       "",
			riskTolerance: "",
			setupBag:      func(bag.SharedBag) {},
			wantErr:       true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			sharedBag := bag.NewSharedBag()
			c.setupBag(sharedBag)

			manager, err := NewProfileManager(testFS, ".", sharedBag)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			profile, err := manager.LoadProfile(ctx, c.country, c.riskTolerance)

			if c.wantErr {
				require.Error(t, err)
				assert.Nil(t, profile)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, profile)

			// Verify the profile was stored in SharedBag
			bagProfile, ok := sharedBag.Get(keys.KProfile)
			assert.True(t, ok)
			assert.Equal(t, profile, bagProfile)

			// Verify expected source
			switch c.expectSource {
			case "country":
				assert.Equal(t, "income", profile.InvestmentStyle)
				assert.Equal(t, "FR", profile.RegionalContext.Country)
			case "global":
				assert.Equal(t, "balanced", profile.InvestmentStyle)
				assert.Equal(t, "", profile.RegionalContext.Country)
			case "bag":
				assert.Equal(t, "growth", profile.InvestmentStyle)
				assert.Equal(t, "comprehensive", profile.ResearchDepth)
			}
		})
	}
}

func TestProfileManager_SharedBagCache(t *testing.T) {
	t.Parallel()

	testFS := newTestFS()
	testFS.MapFS["investment_profiles/defaults/global/moderate.yaml"] = &fstest.MapFile{
		Data: []byte(`investment_style: 'balanced'
research_depth: 'intermediate'
risk_tolerance: 'moderate'

preferred_assets:
  - 'equities'
  - 'bonds'

avoided_sectors: []

esg_criteria:
  esg_importance: 'low'
  esg_focus: []
  exclusion_criteria: []

regional_context:
  country: ''
  language: 'en'
  currency: 'USD'
  timezone: 'UTC'

profile_version: '1.0'
source: 'default_global'
`),
	}

	sharedBag := bag.NewSharedBag()
	manager, err := NewProfileManager(testFS, ".", sharedBag)
	require.NoError(t, err)

	ctx := context.Background()

	// First call should load from filesystem
	profile1, err := manager.LoadProfile(ctx, "US", "moderate")
	require.NoError(t, err)
	assert.Equal(t, "balanced", profile1.InvestmentStyle)

	// Verify it's in the SharedBag
	bagProfile, ok := sharedBag.Get(keys.KProfile)
	require.True(t, ok)
	assert.Equal(t, profile1, bagProfile)

	// Second call should return from SharedBag (not filesystem)
	// Since SharedBag has the profile, it should return that regardless of params
	profile2, err := manager.LoadProfile(ctx, "US", "moderate")
	require.NoError(t, err)
	assert.Equal(t, profile1, profile2) // Should be same instance from SharedBag
}

func TestProfileManager_Interface(t *testing.T) {
	t.Parallel()

	// Verify ProfileManager implements Manager interface
	var _ Manager = &ProfileManager{}

	manager, err := NewProfileManager(newTestFS(), ".", bag.NewSharedBag())
	require.NoError(t, err)

	// Verify interface methods exist
	assert.NotNil(t, manager.LoadProfile)
	assert.NotNil(t, manager.SaveProfile)
}

func TestProfileManager_SaveProfile(t *testing.T) {
	t.Parallel()

	testFS := newTestFS()

	sharedBag := bag.NewSharedBag()
	manager, err := NewProfileManager(testFS, ".", sharedBag)
	require.NoError(t, err)

	profile := &models.InvestmentProfile{
		InvestmentStyle: "growth",
		ResearchDepth:   "comprehensive",
		RegionalContext: models.RegionalInvestmentContext{
			Country:  "US",
			Language: "en",
			Currency: "USD",
			Timezone: "America/New_York",
		},
		CustomRequirements: []string{"ESG focused", "Tax efficient"},
	}

	ctx := context.Background()

	// Test SaveProfile
	filename := "testuser_aggressive.yaml"
	err = manager.SaveProfile(ctx, profile, filename)
	require.NoError(t, err)

	// Verify file was created in testFS
	expectedPath := filepath.Join("profiles", filename)
	data, err := testFS.ReadFile(expectedPath)
	require.NoError(t, err, "Profile file should exist in testFS")

	// Verify file contents
	var savedProfile models.InvestmentProfile
	err = yaml.Unmarshal(data, &savedProfile)
	require.NoError(t, err)

	assert.Equal(t, profile.InvestmentStyle, savedProfile.InvestmentStyle)
	assert.Equal(t, profile.ResearchDepth, savedProfile.ResearchDepth)
	assert.Equal(t, profile.RegionalContext.Country, savedProfile.RegionalContext.Country)
	assert.Equal(t, profile.CustomRequirements, savedProfile.CustomRequirements)
}
