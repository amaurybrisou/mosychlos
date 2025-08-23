package profile

import (
	"context"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Manager handles investment profile loading and management
type Manager interface {
	LoadProfile(ctx context.Context, country, riskTolerance string) (*models.InvestmentProfile, error)
	SaveProfile(ctx context.Context, profile *models.InvestmentProfile, filename string) error
}

// ProfileManager implements the Manager interface
type ProfileManager struct {
	fs        fs.FS
	sharedBag bag.SharedBag
	configDir string // For saving user profiles
}

// NewProfileManager creates a new profile manager
func NewProfileManager(filesystem fs.FS, configDir string, sharedBag bag.SharedBag) (*ProfileManager, error) {
	if filesystem == nil {
		return nil, fmt.Errorf("filesystem cannot be nil")
	}

	if configDir == "" {
		return nil, fmt.Errorf("dataDir cannot be empty")
	}

	return &ProfileManager{
		fs:        filesystem,
		sharedBag: sharedBag,
		configDir: configDir,
	}, nil
}

// LoadProfile loads a profile for specific country and risk tolerance
// Falls back to global profile if country-specific profile is not found
// Follows the same caching pattern as tools
func (pm *ProfileManager) LoadProfile(ctx context.Context, country, riskTolerance string) (*models.InvestmentProfile, error) {
	if country == "" || riskTolerance == "" {
		return nil, fmt.Errorf("country and risk tolerance are required")
	}

	// Check SharedBag first (like tools pattern)
	if pm.sharedBag != nil {
		if cached, exists := pm.sharedBag.Get(bag.KProfile); exists {
			if profile, ok := cached.(*models.InvestmentProfile); ok {
				return profile, nil
			}
		}
	}

	// Try country-specific profile first
	profile, err := pm.loadFromPath(pm.getProfilePath(country, riskTolerance))
	if err == nil {
		// Cache successful load
		if pm.sharedBag != nil {
			pm.sharedBag.Set(bag.KProfile, profile)
		}
		return profile, nil
	}

	// Fallback to global profile
	profile, err = pm.loadFromPath(pm.getGlobalProfilePath(riskTolerance))
	if err != nil {
		return nil, fmt.Errorf("failed to load profile for %s/%s and global fallback: %w", country, riskTolerance, err)
	}

	// Cache the global profile under the requested key
	if pm.sharedBag != nil {
		pm.sharedBag.Set(bag.KProfile, profile)
	}
	return profile, nil
}

// loadFromPath loads and parses a profile from filesystem path
func (pm *ProfileManager) loadFromPath(profilePath string) (*models.InvestmentProfile, error) {
	data, err := pm.fs.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file %s: %w", profilePath, err)
	}

	return pm.parseProfileData(data)
}

// parseProfileData parses YAML data into an InvestmentProfile
func (pm *ProfileManager) parseProfileData(data []byte) (*models.InvestmentProfile, error) {
	var profile models.InvestmentProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile YAML: %w", err)
	}
	return &profile, nil
}

// getProfilePath constructs the path for a country/risk combination
func (pm *ProfileManager) getProfilePath(country, riskTolerance string) string {
	return filepath.Join(pm.configDir, "investment_profiles", "defaults", country, riskTolerance+".yaml")
}

// getGlobalProfilePath constructs the path for a global profile
func (pm *ProfileManager) getGlobalProfilePath(riskTolerance string) string {
	return filepath.Join(pm.configDir, "investment_profiles", "defaults", "global", riskTolerance+".yaml")
}

// SaveProfile saves a user profile to DataDir/profiles
func (pm *ProfileManager) SaveProfile(ctx context.Context, profile *models.InvestmentProfile, filename string) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Ensure profiles directory exists
	profilesDir := filepath.Join(pm.configDir, "profiles")
	if err := pm.fs.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Add .yaml extension if not present
	if filepath.Ext(filename) == "" {
		filename += ".yaml"
	}

	// Serialize to YAML
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to serialize profile: %w", err)
	}

	// Write to file
	profilePath := filepath.Join(profilesDir, filename)
	if err := pm.fs.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	// Update SharedBag with saved profile
	if pm.sharedBag != nil {
		pm.sharedBag.Set(bag.KProfile, profile)
	}

	return nil
}
