package errorcode

const (
	// CodeNotFound 表示资源未找到。Service 层收到的错误中若是这样的 payload 则表示是资源未找到，而非链码运行出错
	CodeNotFound = "~NOTFOUND~"
	// CodeNotImplemented 是个在这个项目中约定俗成的代号。Service 层收到错误中若是这样的 payload 则表示是暂时未实现的功能而非错误。
	CodeNotImplemented = "~NOTIMPLEMENTED~"
)
