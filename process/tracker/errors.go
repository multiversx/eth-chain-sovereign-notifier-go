package tracker

import (
	"errors"
)

var errNilClient = errors.New("nil eth client provided")

var errInvalidMaxCacheSize = errors.New("invalid block cache size")

var errInvalidMinConfirmations = errors.New("invalid number of block confirmation provided")
