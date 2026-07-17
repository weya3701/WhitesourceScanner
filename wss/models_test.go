package wss

import (
	"encoding/json"
	"testing"
)

func TestProcessStatusAcceptsNumericOrStringContextID(t *testing.T) {
	for _, body := range []string{
		`{"asyncProcessStatus":{"contextId":123,"status":"SUCCESS"}}`,
		`{"asyncProcessStatus":{"contextId":"123","status":"SUCCESS"}}`,
	} {
		var response ProcessStatusResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			t.Fatalf("failed to decode %s: %v", body, err)
		}
	}
}
