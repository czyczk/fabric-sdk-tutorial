package controller

import "strings"

// "+" sign in URL should be kept unchanged (instead of being changed into a " ") for Base64 encoded strings.
func processBase64FromURLQuery(parameterValue string) string {
	return strings.ReplaceAll(parameterValue, " ", "+")
}
