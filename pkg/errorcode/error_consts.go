package errorcode

import "fmt"

const (
	// CodeNotFound 表示资源未找到。Service 层收到的错误中若是这样的错误信息则表示是资源未找到，而非链码运行出错。
	CodeNotFound = "~NOTFOUND~"
	// CodeForbidden 表示参数被理解，但无权进行操作。Service 层收到的错误中若是这样的错误信息则表示是操作权限的问题，而非链码运行出错。
	CodeForbidden = "~FORBIDDEN~"
	// CodeNotImplemented 是个在这个项目中约定俗成的代号。Service 层收到错误中若是这样的错误信息则表示是暂时未实现的功能而非链码运行出错。
	CodeNotImplemented = "~NOTIMPLEMENTED~"
)

// ErrorNotFound 为使用了 `CodeNotFound` 的 error 实例
var ErrorNotFound = fmt.Errorf(CodeNotFound)

// ErrorForbidden 为使用了 `CodeForbidden` 的 error 实例
var ErrorForbidden = fmt.Errorf(CodeForbidden)

// ErrorNotImplemented 为使用了 `CodeNotImplemented` 的 error 实例
var ErrorNotImplemented = fmt.Errorf(CodeNotImplemented)
