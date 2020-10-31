package controller

// GeneralResponse is a response that should be produced any sent via an endpoint handler.
type GeneralResponse struct {
	errors ParameterErrorList
	msg    string
}

// NewFromErrors fills a GeneralResponse with errors.
func (gr *GeneralResponse) NewFromErrors(errors *ParameterErrorList) {
	gr.errors = *errors
}

// NewFromMsg fills a GeneralResponse with a string message.
func (gr *GeneralResponse) NewFromMsg(msg string) {
	gr.msg = msg
}

// ToMap converts this struct to a map.
func (gr *GeneralResponse) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"errors": gr.errors,
		"msg":    gr.msg,
	}
}
