package collection

import (
	"errors"
)

var (
	// ErrWrongPeriod -
	ErrWrongPeriod = errors.New("ErrWrongPeriod")

	// ErrUnsupportedType -
	ErrUnsupportedType = errors.New("ErrUnsupportedType")

	// ErrLengthLowerThanZero -
	ErrLengthLowerThanZero = errors.New("ErrLengthLowerThanZero")

	// ErrUnsupportedDateFormat -
	ErrUnsupportedDateFormat = errors.New("ErrUnsupportedDateFormat")
)
