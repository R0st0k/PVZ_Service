package sl

import (
	"errors"
	"testing"
)

func TestErr(t *testing.T) {
	testError := errors.New("test error")

	attr := Err(testError)

	if attr.Key != "error" {
		t.Errorf("Expected attribute key to be 'error', got '%s'", attr.Key)
	}

	value := attr.Value.String()
	if value != testError.Error() {
		t.Errorf("Expected attribute value to be '%s', got '%s'", testError.Error(), value)
	}
}
