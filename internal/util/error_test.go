package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr temporarily redirects os.Stderr to capture output.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestExitError_Error(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		err     error
		want    string
	}{
		{
			name:    "message only without wrapped error",
			code:    1,
			message: "something failed",
			err:     nil,
			want:    "something failed",
		},
		{
			name:    "message with wrapped error",
			code:    2,
			message: "operation failed",
			err:     errors.New("underlying cause"),
			want:    "operation failed: underlying cause",
		},
		{
			name:    "empty message without error",
			code:    1,
			message: "",
			err:     nil,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExitError{Code: tt.code, Message: tt.message, Err: tt.err}
			if got := e.Error(); got != tt.want {
				t.Errorf("ExitError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExitError_Unwrap(t *testing.T) {
	t.Run("returns wrapped error", func(t *testing.T) {
		underlying := errors.New("cause")
		e := &ExitError{Code: 1, Message: "msg", Err: underlying}
		if got := e.Unwrap(); got != underlying {
			t.Errorf("Unwrap() = %v, want %v", got, underlying)
		}
	})

	t.Run("returns nil when no wrapped error", func(t *testing.T) {
		e := &ExitError{Code: 1, Message: "msg"}
		if got := e.Unwrap(); got != nil {
			t.Errorf("Unwrap() = %v, want nil", got)
		}
	})
}

func TestExitError_ImplementsError(t *testing.T) {
	var _ error = &ExitError{}
}

func TestExitError_ErrorsIs(t *testing.T) {
	target := errors.New("target error")
	e := &ExitError{Code: 1, Message: "wrapper", Err: target}

	if !errors.Is(e, target) {
		t.Error("errors.Is() should return true for wrapped error")
	}
}

func TestNewExitError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		err     error
	}{
		{"basic exit error", 1, "test error", nil},
		{"with wrapped error", 2, "wrapped", errors.New("cause")},
		{"exit code 0", 0, "success", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExitError(tt.code, tt.message, tt.err)
			if got == nil {
				t.Fatal("NewExitError() returned nil")
			}
			if got.Code != tt.code {
				t.Errorf("ExitError.Code = %d, want %d", got.Code, tt.code)
			}
			if got.Message != tt.message {
				t.Errorf("ExitError.Message = %q, want %q", got.Message, tt.message)
			}
			if got.Err != tt.err {
				t.Errorf("ExitError.Err = %v, want %v", got.Err, tt.err)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			format: "something went wrong",
			want:   "Error: something went wrong\n",
		},
		{
			name:   "formatted message",
			format: "failed to open %s: %v",
			args:   []interface{}{"file.txt", fmt.Errorf("not found")},
			want:   "Error: failed to open file.txt: not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStderr(t, func() {
				PrintError(tt.format, tt.args...)
			})
			if got != tt.want {
				t.Errorf("PrintError() output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintWarning(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple warning",
			format: "this is deprecated",
			want:   "Warning: this is deprecated\n",
		},
		{
			name:   "formatted warning",
			format: "value %d is out of range",
			args:   []interface{}{100},
			want:   "Warning: value 100 is out of range\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStderr(t, func() {
				PrintWarning(tt.format, tt.args...)
			})
			if got != tt.want {
				t.Errorf("PrintWarning() output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintDebug(t *testing.T) {
	tests := []struct {
		name    string
		debug   bool
		format  string
		args    []interface{}
		want    string
		wantOut bool
	}{
		{
			name:    "debug enabled prints message",
			debug:   true,
			format:  "request to %s",
			args:    []interface{}{"api.example.com"},
			want:    "[DEBUG] request to api.example.com\n",
			wantOut: true,
		},
		{
			name:    "debug disabled prints nothing",
			debug:   false,
			format:  "hidden message",
			wantOut: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStderr(t, func() {
				PrintDebug(tt.debug, tt.format, tt.args...)
			})
			if tt.wantOut {
				if got != tt.want {
					t.Errorf("PrintDebug() output = %q, want %q", got, tt.want)
				}
			} else {
				if strings.TrimSpace(got) != "" {
					t.Errorf("PrintDebug() should produce no output when debug=false, got %q", got)
				}
			}
		})
	}
}
