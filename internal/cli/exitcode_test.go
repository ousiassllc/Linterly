package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitError_Error(t *testing.T) {
	e := &ExitError{Code: 2, Message: "something went wrong"}
	assert.Equal(t, "something went wrong", e.Error())
}

func TestExitError_ErrorsAs(t *testing.T) {
	var orig error = &ExitError{Code: 2, Message: "test"}

	var target *ExitError
	assert.True(t, errors.As(orig, &target))
	assert.Equal(t, 2, target.Code)
	assert.Equal(t, "test", target.Message)
}

func TestNewRuntimeError(t *testing.T) {
	e := NewRuntimeError("file not found: %s", "foo.yml")
	assert.Equal(t, ExitRuntimeError, e.Code)
	assert.Equal(t, "file not found: foo.yml", e.Message)
}

func TestNewViolationError(t *testing.T) {
	e := NewViolationError()
	assert.Equal(t, ExitViolation, e.Code)
	assert.Equal(t, "violations found", e.Message)
}

func TestExitError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	e := &ExitError{Code: 2, Message: "wrapped", Cause: cause}
	assert.Equal(t, cause, e.Unwrap())
}

func TestExitError_Unwrap_NilCause(t *testing.T) {
	e := &ExitError{Code: 2, Message: "no cause"}
	assert.Nil(t, e.Unwrap())
}
