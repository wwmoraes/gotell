package gotell

import (
	"path"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// FunctionInfo contains details about a function call.
type FunctionInfo struct {
	Package      string
	Filename     string
	Directory    string
	Filepath     string
	FunctionName string
	LineNumber   int
}

// GetFunctionInfo reports information about function invocations on the calling
// goroutine's stack. It parses data from runtime.Caller and runtime.FuncForPC
// to provide a structured result.
//
// Skip equals zero means the caller of this function.
func GetFunctionInfo(skip int) *FunctionInfo {
	//nolint:exhaustruct // its filled down below
	info := FunctionInfo{}

	var pc uintptr

	pc, info.Filepath, info.LineNumber, _ = runtime.Caller(skip + 1)
	info.Directory, info.Filename = path.Split(info.Filepath)

	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	parts, info.FunctionName = parts[:len(parts)-1], parts[len(parts)-1]

	if parts[len(parts)-1][0] == '(' {
		parts, info.FunctionName = parts[:len(parts)-1], parts[len(parts)-1]+"."+info.FunctionName
	}

	info.Package = strings.Join(parts, ".")

	return &info
}

// FunctionAttributes is a convenience function that runs GetFunctionInfo then
// FunctionInfoAttributes.
//
// Skip equals zero means the caller of this function.
func FunctionAttributes(skip int) []attribute.KeyValue {
	return FunctionInfoAttributes(GetFunctionInfo(skip + 1))
}

// FunctionInfoAttributes converts the information from FunctionInfo into
// OpenTelemetry semantic attributes.
func FunctionInfoAttributes(info *FunctionInfo) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("code.function", info.FunctionName),
		attribute.String("code.filepath", info.Filepath),
		attribute.Int("code.lineno", info.LineNumber),
		attribute.String("code.namespace", info.Package),
	}
}
