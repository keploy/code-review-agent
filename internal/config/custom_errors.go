package config
import (
	"errors"
)

var ErrTokenLimitExceeded = errors.New("token limit exceeded")