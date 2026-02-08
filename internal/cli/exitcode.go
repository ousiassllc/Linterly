package cli

import "fmt"

// 終了コード定数。
const (
	// ExitOK は正常終了（違反なし）を示す。直接参照はされないが、
	// 終了コード体系のドキュメントとして定義している。
	ExitOK           = 0
	ExitViolation    = 1
	ExitRuntimeError = 2
)

// ExitError はプロセスの終了コードを伴うエラー。
type ExitError struct {
	Code    int
	Message string
	Cause   error
}

func (e *ExitError) Error() string {
	return e.Message
}

// Unwrap は原因エラーを返す。errors.Is / errors.As でのチェーン辿りに対応する。
func (e *ExitError) Unwrap() error {
	return e.Cause
}

// NewRuntimeError は Code=2（実行エラー）の ExitError を返す。
func NewRuntimeError(format string, args ...any) *ExitError {
	return &ExitError{
		Code:    ExitRuntimeError,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewViolationError は Code=1（違反あり）の ExitError を返す。
func NewViolationError() *ExitError {
	return &ExitError{
		Code:    ExitViolation,
		Message: "violations found",
	}
}
