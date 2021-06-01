package controller

import (
	"strconv"
	"strings"
	"time"
)

// ParameterErrorList contains a list of human-readable errors about parameters.
type ParameterErrorList []string

// AppendIfEmptyOrBlankSpaces appends the error message specified if `str` is empty or contains only blank spaces.
//
// Parameters:
//   the string to be checked
//   the error message to append
//
// Returns:
//   the trimmed string
func (pel *ParameterErrorList) AppendIfEmptyOrBlankSpaces(str string, errMsg string) string {
	if str = strings.TrimSpace(str); str == "" {
		*pel = append(*pel, errMsg)
	}

	return str
}

// AppendIfNotInt appends the error message specified if `str` is not an int.
//
// Parameters:
//   the string to be checked
//   the error message to append
//
// Returns:
//   the parsed int or 0 if there's error
func (pel *ParameterErrorList) AppendIfNotInt(str string, errMsg string) int {
	intResult, err := strconv.Atoi(str)
	if err != nil {
		*pel = append(*pel, errMsg)
	}

	return intResult
}

// AppendIfNotPositiveInt appends the error message specified if `str` is not a positive int.
//
// Parameters:
//   the string to be checked
//   the error message to append
//
// Returns:
//   the parsed int or 0 if it can't be parsed as int
func (pel *ParameterErrorList) AppendIfNotPositiveInt(str string, errMsg string) int {
	intResult, err := strconv.Atoi(str)
	if err != nil {
		*pel = append(*pel, errMsg)
		return 0
	}

	if intResult < 0 {
		*pel = append(*pel, errMsg)
	}

	return intResult
}

// AppendIfNotUint appends the error message specified if `str` is not an uint.
//
// Parameters:
//   the string to be checked
//   the error message to append
//
// Returns:
//   the parsed uint or 0 if there's error
func (pel *ParameterErrorList) AppendIfNotUint(str string, errMsg string) uint {
	intResult, err := strconv.Atoi(str)
	if err != nil || intResult < 0 {
		*pel = append(*pel, errMsg)
	}

	return uint(intResult)
}

// AppendIfNotBool appends the error message specified if `str` is not a bool.
//
// Parameters:
//   the string to be checked
//   the error message to append
//
// Returns:
//   the parsed bool
func (pel *ParameterErrorList) AppendIfNotBool(str string, errMsg string) bool {
	boolResult, err := strconv.ParseBool(str)
	if err != nil {
		*pel = append(*pel, errMsg)
	}

	return boolResult
}

// AppendIfNotTime appends the error message specified if `str` does not represent a valid RFC3339 time.
//
// Parameters:
//   the string to be checked
//   the error message to append
// Returns:
//   the parsed time
func (pel *ParameterErrorList) AppendIfNotTime(str string, errMsg string) time.Time {
	timeResult, err := time.Parse(time.RFC3339, str)
	if err != nil {
		*pel = append(*pel, errMsg)
	}

	return timeResult
}
