package filesystem

import (
	datastorage "schneider.vip/retryspool/storage/data"
)

// Factory implements datastorage.Factory for filesystem storage
type Factory struct {
	basePath string
}

// NewFactory creates a new filesystem data storage factory
func NewFactory(basePath string) *Factory {
	return &Factory{
		basePath: basePath,
	}
}

// Create creates a new filesystem data storage backend
func (f *Factory) Create() (datastorage.Backend, error) {
	return NewBackend(f.basePath)
}

// Name returns the factory name
func (f *Factory) Name() string {
	return "filesystem-data"
}
