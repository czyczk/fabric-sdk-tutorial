package bcao

import (
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/pkg/errors"
)

// GetClassifiedError is a general error handler that converts some errors returned from the chaincode to the predefined errors.
func GetClassifiedError(chaincodeFcn string, err error) error {
	if err == nil {
		return nil
	} else if strings.HasSuffix(err.Error(), errorcode.CodeForbidden) {
		return errorcode.ErrorForbidden
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotFound) {
		return errorcode.ErrorNotFound
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
		return errorcode.ErrorNotImplemented
	} else {
		return errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
	}
}
