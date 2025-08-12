package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrProfileNotFound indicates that a profile was not found
	ErrProfileNotFound = errors.New("profile not found")

	// ErrProfileAlreadyExists indicates that a profile with the same name already exists
	ErrProfileAlreadyExists = errors.New("profile already exists")

	// ErrInvalidProfileName indicates an invalid profile name was provided
	ErrInvalidProfileName = errors.New("invalid profile name")

	// ErrEmptyProfileName indicates an empty profile name was provided
	ErrEmptyProfileName = errors.New("profile name cannot be empty")

	// ErrInvalidAllocation indicates an invalid allocation percentage
	ErrInvalidAllocation = errors.New("invalid allocation percentage")

	// ErrTotalAllocationMismatch indicates the total allocation doesn't equal 100%
	ErrTotalAllocationMismatch = errors.New("total allocation must equal 100%")

	// ErrEmptyAssetName indicates an empty asset name
	ErrEmptyAssetName = errors.New("asset name cannot be empty")

	// ErrEmptyProfileGoal indicates an empty profile goal/description
	ErrEmptyProfileGoal = errors.New("profile goal cannot be empty")

	// ErrNoAssetsAllocated indicates no assets were allocated in the profile
	ErrNoAssetsAllocated = errors.New("profile must have at least one allocation")

	// ErrFailedToLoadProfile indicates a profile couldn't be loaded
	ErrFailedToLoadProfile = errors.New("failed to load profile")

	// ErrFailedToSaveProfile indicates a profile couldn't be saved
	ErrFailedToSaveProfile = errors.New("failed to save profile")

	// ErrNoActiveProfile indicates no active profile is set
	ErrNoActiveProfile = errors.New("no active profile set")

	// ErrFailedToSetActiveProfile indicates active profile couldn't be set
	ErrFailedToSetActiveProfile = errors.New("failed to set active profile")
)

// ProfileNotFoundError returns a formatted error for a specific profile
func ProfileNotFoundError(name string) error {
	return fmt.Errorf("%w: %s", ErrProfileNotFound, name)
}

// ProfileAlreadyExistsError returns a formatted error for a specific profile
func ProfileAlreadyExistsError(name string) error {
	return fmt.Errorf("%w: %s", ErrProfileAlreadyExists, name)
}

// InvalidAllocationError returns a formatted error for an invalid allocation
func InvalidAllocationError(asset string, percentage int) error {
	return fmt.Errorf("%w for %s: %d%%", ErrInvalidAllocation, asset, percentage)
}

// TotalAllocationMismatchError returns a formatted error for a total allocation mismatch
func TotalAllocationMismatchError(total int) error {
	return fmt.Errorf("%w: got %d%%", ErrTotalAllocationMismatch, total)
}

// FailedToLoadProfileError returns a formatted error when loading fails
func FailedToLoadProfileError(name string, err error) error {
	return fmt.Errorf("%w %s: %v", ErrFailedToLoadProfile, name, err)
}

// FailedToSaveProfileError returns a formatted error when saving fails
func FailedToSaveProfileError(name string, err error) error {
	return fmt.Errorf("%w %s: %v", ErrFailedToSaveProfile, name, err)
}

// IsProfileNotFound checks if the error is a profile not found error
func IsProfileNotFound(err error) bool {
	return errors.Is(err, ErrProfileNotFound)
}

// IsProfileAlreadyExists checks if the error is a profile already exists error
func IsProfileAlreadyExists(err error) bool {
	return errors.Is(err, ErrProfileAlreadyExists)
}

// IsInvalidAllocation checks if the error is an invalid allocation error
func IsInvalidAllocation(err error) bool {
	return errors.Is(err, ErrInvalidAllocation)
}

// IsTotalAllocationMismatch checks if the error is a total allocation mismatch error
func IsTotalAllocationMismatch(err error) bool {
	return errors.Is(err, ErrTotalAllocationMismatch)
}

// IsNoActiveProfile checks if the error indicates no active profile
func IsNoActiveProfile(err error) bool {
	return errors.Is(err, ErrNoActiveProfile)
}
