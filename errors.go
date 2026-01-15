package boxpacker3

import (
	"errors"
)

var (
	ErrInvalidDimension  = errors.New("invalid dimension")
	ErrInvalidWeight     = errors.New("invalid weight")
	ErrInnerExceedsOuter = errors.New("inner dimensions exceed outer dimensions")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInvalidSetting    = errors.New("invalid setting")
	ErrInvalidAxis       = errors.New("invalid vertical axis")
	ErrInvalidClass      = errors.New("invalid goods class")
)
