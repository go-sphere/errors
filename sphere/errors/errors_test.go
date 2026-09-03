package errors_test

import (
	"testing"

	"github.com/go-sphere/errors/sphere/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestError_GettersAndNilSafety(t *testing.T) {
	var nilErr *errors.Error
	if nilErr.GetStatus() != 0 {
		t.Errorf("expected 0 for nil receiver GetStatus")
	}
	if nilErr.GetReason() != "" {
		t.Errorf("expected empty string for nil receiver GetReason")
	}
	if nilErr.GetMessage() != "" {
		t.Errorf("expected empty string for nil receiver GetMessage")
	}

	errObj := &errors.Error{
		Status:  404,
		Reason:  "NOT_FOUND",
		Message: "Resource not found",
	}
	if errObj.GetStatus() != 404 {
		t.Errorf("expected 404, got %d", errObj.GetStatus())
	}
	if errObj.GetReason() != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", errObj.GetReason())
	}
	if errObj.GetMessage() != "Resource not found" {
		t.Errorf("expected Resource not found, got %s", errObj.GetMessage())
	}
}

func TestError_ProtoRoundTrip(t *testing.T) {
	orig := &errors.Error{
		Status:  400,
		Reason:  "BAD_REQUEST",
		Message: "Invalid input parameter",
	}

	data, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}

	var parsed errors.Error
	if err := proto.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if parsed.GetStatus() != orig.GetStatus() ||
		parsed.GetReason() != orig.GetReason() ||
		parsed.GetMessage() != orig.GetMessage() {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", &parsed, orig)
	}
}

func TestErrorExtensions(t *testing.T) {
	enumOpts := &descriptorpb.EnumOptions{}
	proto.SetExtension(enumOpts, errors.E_DefaultStatus, int32(500))

	if !proto.HasExtension(enumOpts, errors.E_DefaultStatus) {
		t.Fatal("expected DefaultStatus extension")
	}
	val := proto.GetExtension(enumOpts, errors.E_DefaultStatus).(int32)
	if val != 500 {
		t.Errorf("expected 500, got %d", val)
	}

	valOpts := &descriptorpb.EnumValueOptions{}
	errVal := &errors.Error{Status: 403, Reason: "FORBIDDEN", Message: "Access denied"}
	proto.SetExtension(valOpts, errors.E_Options, errVal)

	if !proto.HasExtension(valOpts, errors.E_Options) {
		t.Fatal("expected Options extension")
	}
	got := proto.GetExtension(valOpts, errors.E_Options).(*errors.Error)
	if got.GetStatus() != 403 || got.GetReason() != "FORBIDDEN" || got.GetMessage() != "Access denied" {
		t.Errorf("unexpected error option: %+v", got)
	}
}
