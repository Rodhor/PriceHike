package domain

import "errors"

var (
	ErrNoDatasetLoaded   = errors.New("no dataset loaded")
	ErrInvalidRule       = errors.New("invalid rule")
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrInvalidSettings   = errors.New("invalid settings")
)
