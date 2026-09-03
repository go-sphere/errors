package errors_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/go-sphere/errors/sphere/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TestError_NilReceiverMethods tests all method calls on a nil *errors.Error receiver.
func TestError_NilReceiverMethods(t *testing.T) {
	var nilErr *errors.Error

	// Getters
	if nilErr.GetStatus() != 0 {
		t.Errorf("expected 0, got %d", nilErr.GetStatus())
	}
	if nilErr.GetReason() != "" {
		t.Errorf("expected empty string, got %q", nilErr.GetReason())
	}
	if nilErr.GetMessage() != "" {
		t.Errorf("expected empty string, got %q", nilErr.GetMessage())
	}

	// String and Descriptor
	if nilErr.String() == "" {
		t.Errorf("expected non-empty proto string for nil receiver")
	}
	raw, indices := nilErr.Descriptor()
	if len(raw) == 0 || len(indices) == 0 {
		t.Errorf("Descriptor() on nil receiver should return valid descriptor info")
	}

	// ProtoReflect on nil receiver
	ref := nilErr.ProtoReflect()
	if ref.IsValid() {
		t.Errorf("ProtoReflect on nil receiver should not be valid")
	}

	// Reset on non-nil receiver succeeds
	resetTarget := &errors.Error{Status: 500, Reason: "FAIL", Message: "msg"}
	resetTarget.Reset()
	if resetTarget.GetStatus() != 0 || resetTarget.GetReason() != "" {
		t.Errorf("Reset did not clear message: %+v", resetTarget)
	}

	// Proto standard operations on nil receiver
	if proto.Size(nilErr) != 0 {
		t.Errorf("proto.Size(nil) should be 0, got %d", proto.Size(nilErr))
	}
	cloned := proto.Clone(nilErr)
	if cloned != nil && !proto.Equal(cloned, nilErr) {
		t.Errorf("proto.Clone(nil) should be equal to nil, got %v", cloned)
	}
	if !proto.Equal(nilErr, nilErr) {
		t.Error("proto.Equal(nil, nil) should be true")
	}
	if proto.Equal(nilErr, &errors.Error{}) {
		t.Error("proto.Equal(nil, &Error{}) should be false")
	}

	data, err := proto.Marshal(nilErr)
	if err != nil {
		t.Errorf("proto.Marshal(nil) error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("proto.Marshal(nil) should produce empty bytes, got %d bytes", len(data))
	}

	// Nil descriptor extensions
	var nilEnumOpts *descriptorpb.EnumOptions
	if proto.HasExtension(nilEnumOpts, errors.E_DefaultStatus) {
		t.Error("HasExtension on nil EnumOptions should be false")
	}

	var nilValOpts *descriptorpb.EnumValueOptions
	if proto.HasExtension(nilValOpts, errors.E_Options) {
		t.Error("HasExtension on nil EnumValueOptions should be false")
	}
}

// TestError_BoundaryAndWireRoundtrips tests wire serialization roundtrips with
// boundary statuses, unicode, and large messages.
func TestError_BoundaryAndWireRoundtrips(t *testing.T) {
	testCases := []struct {
		name    string
		status  int32
		reason  string
		message string
	}{
		{"zero_all", 0, "", ""},
		{"status_100", 100, "CONTINUE", "Continue request"},
		{"status_200", 200, "OK", "Success"},
		{"status_400", 400, "BAD_REQUEST", "Malformed body"},
		{"status_404", 404, "NOT_FOUND", "Resource missing"},
		{"status_500", 500, "INTERNAL_ERROR", "Server crash"},
		{"status_599", 599, "NETWORK_TIMEOUT", "Gateway timeout variant"},
		{"negative_status", -1, "NEGATIVE", "Negative status code"},
		{"large_status", 99999, "CUSTOM_ERR", "High custom error"},
		{"unicode_message", 403, "权限不足", "当前用户无权访问此资源 🔒"},
		{"special_chars", 422, "VALIDATION_FAILED", "Line1\nLine2\tTab\"Quote'<script>alert(1)</script>"},
		{"large_payload", 500, "LARGE_ERR", strings.Repeat("A very long error message detailing stack traces. ", 500)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &errors.Error{
				Status:  tc.status,
				Reason:  tc.reason,
				Message: tc.message,
			}

			data, err := proto.Marshal(orig)
			if err != nil {
				t.Fatalf("proto.Marshal failed: %v", err)
			}

			var parsed errors.Error
			if err := proto.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("proto.Unmarshal failed: %v", err)
			}

			if !proto.Equal(orig, &parsed) {
				t.Errorf("roundtrip equality failed: got %+v, want %+v", &parsed, orig)
			}
			if parsed.GetStatus() != tc.status {
				t.Errorf("status: got %d, want %d", parsed.GetStatus(), tc.status)
			}
			if parsed.GetReason() != tc.reason {
				t.Errorf("reason: got %q, want %q", parsed.GetReason(), tc.reason)
			}
			if parsed.GetMessage() != tc.message {
				t.Errorf("message: got %q, want %q", parsed.GetMessage(), tc.message)
			}
		})
	}
}

// TestError_ConcurrentRoundtripStress tests concurrent error creation,
// extension manipulation, marshaling, and unmarshaling under race detection.
func TestError_ConcurrentRoundtripStress(t *testing.T) {
	const goroutines = 50
	const iterations = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				status := int32(400 + (id+i)%100)
				reason := fmt.Sprintf("ERR_%d_%d", id, i)
				msg := fmt.Sprintf("Error detail for worker %d at iteration %d", id, i)

				errObj := &errors.Error{
					Status:  status,
					Reason:  reason,
					Message: msg,
				}

				// Direct marshaling roundtrip
				data, err := proto.Marshal(errObj)
				if err != nil {
					t.Errorf("goroutine %d: proto.Marshal error: %v", id, err)
					return
				}

				var parsed errors.Error
				if err := proto.Unmarshal(data, &parsed); err != nil {
					t.Errorf("goroutine %d: proto.Unmarshal error: %v", id, err)
					return
				}

				if !proto.Equal(errObj, &parsed) {
					t.Errorf("goroutine %d: roundtrip mismatch", id)
					return
				}

				// Extension roundtrip
				valOpts := &descriptorpb.EnumValueOptions{}
				proto.SetExtension(valOpts, errors.E_Options, errObj)

				valData, err := proto.Marshal(valOpts)
				if err != nil {
					t.Errorf("goroutine %d: proto.Marshal EnumValueOptions error: %v", id, err)
					return
				}

				parsedValOpts := &descriptorpb.EnumValueOptions{}
				if err := proto.Unmarshal(valData, parsedValOpts); err != nil {
					t.Errorf("goroutine %d: proto.Unmarshal EnumValueOptions error: %v", id, err)
					return
				}

				if !proto.HasExtension(parsedValOpts, errors.E_Options) {
					t.Errorf("goroutine %d: missing E_Options extension", id)
					return
				}
				extErr := proto.GetExtension(parsedValOpts, errors.E_Options).(*errors.Error)
				if !proto.Equal(errObj, extErr) {
					t.Errorf("goroutine %d: extension Error mismatch", id)
					return
				}
			}
		}(g)
	}

	wg.Wait()
}
