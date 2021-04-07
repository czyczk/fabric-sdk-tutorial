package errorcode

import "fmt"

const (
	// CodeForbidden 表示参数被理解，但无权进行操作。收到的错误中若是这样的错误信息则表示是操作权限的问题，而非链码运行出错。对应 HTTP 状态码的 403。
	CodeForbidden = "~FORBIDDEN~"
	// CodeNotFound 表示资源未找到。收到的错误中若是这样的错误信息则表示是资源未找到，而非链码运行出错。对应 HTTP 状态码的 404。
	CodeNotFound = "~NOTFOUND~"
	// CodeNotImplemented 是个在这个项目中约定俗成的代号。收到错误中若是这样的错误信息则表示是暂时未实现的功能而非链码运行出错。对应 HTTP 状态码的 500。
	CodeNotImplemented = "~NOTIMPLEMENTED~"
	// CodeGatewayTimeout 是个在这个项目中约定俗成的代号。收到错误中若是这样的错误信息则表示是因操作超时引起的。对应 HTTP 状态码的 504。
	CodeGatewayTimeout = "~GATEWAYTIMEOUT~"
)

// ErrorForbidden 为使用了 `CodeForbidden` 的 error 实例
var ErrorForbidden = fmt.Errorf(CodeForbidden)

// ErrorNotFound 为使用了 `CodeNotFound` 的 error 实例
var ErrorNotFound = fmt.Errorf(CodeNotFound)

// ErrorNotImplemented 为使用了 `CodeNotImplemented` 的 error 实例
var ErrorNotImplemented = fmt.Errorf(CodeNotImplemented)

// ErrorGatewayTimeout 为使用了 `CodeGatewayTimeout` 的 error 实例
var ErrorGatewayTimeout = fmt.Errorf(CodeGatewayTimeout)
