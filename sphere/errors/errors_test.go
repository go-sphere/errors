package errors_test

import (
	"testing"

	"github.com/go-sphere/errors/sphere/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestErrorDescriptor(t *testing.T) {
	descriptor := (&errors.Error{}).ProtoReflect().Descriptor()
	if got, want := descriptor.FullName(), protoreflect.FullName("sphere.errors.Error"); got != want {
		t.Fatalf("full name = %q, want %q", got, want)
	}

	tests := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
		kind   protoreflect.Kind
	}{
		{name: "status", number: 1, kind: protoreflect.Int32Kind},
		{name: "reason", number: 2, kind: protoreflect.StringKind},
		{name: "message", number: 3, kind: protoreflect.StringKind},
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			field := descriptor.Fields().ByName(tt.name)
			if field == nil {
				t.Fatalf("field %q not found", tt.name)
			}
			if got := field.Number(); got != tt.number {
				t.Errorf("number = %d, want %d", got, tt.number)
			}
			if got := field.Kind(); got != tt.kind {
				t.Errorf("kind = %s, want %s", got, tt.kind)
			}
		})
	}
}

func TestExtensionDescriptors(t *testing.T) {
	tests := []struct {
		name        string
		extension   protoreflect.ExtensionType
		fullName    protoreflect.FullName
		number      protoreflect.FieldNumber
		kind        protoreflect.Kind
		extendee    protoreflect.FullName
		cardinality protoreflect.Cardinality
	}{
		{
			name: "default_status", extension: errors.E_DefaultStatus,
			fullName: "sphere.errors.default_status", number: 18534200,
			kind: protoreflect.Int32Kind, extendee: "google.protobuf.EnumOptions",
			cardinality: protoreflect.Optional,
		},
		{
			name: "options", extension: errors.E_Options,
			fullName: "sphere.errors.options", number: 18534210,
			kind: protoreflect.MessageKind, extendee: "google.protobuf.EnumValueOptions",
			cardinality: protoreflect.Optional,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor := tt.extension.TypeDescriptor()
			if got := descriptor.FullName(); got != tt.fullName {
				t.Errorf("full name = %q, want %q", got, tt.fullName)
			}
			if got := descriptor.Number(); got != tt.number {
				t.Errorf("number = %d, want %d", got, tt.number)
			}
			if got := descriptor.Kind(); got != tt.kind {
				t.Errorf("kind = %s, want %s", got, tt.kind)
			}
			if got := descriptor.ContainingMessage().FullName(); got != tt.extendee {
				t.Errorf("extendee = %q, want %q", got, tt.extendee)
			}
			if got := descriptor.Cardinality(); got != tt.cardinality {
				t.Errorf("cardinality = %s, want %s", got, tt.cardinality)
			}
		})
	}
}

func TestExtensionsWireRoundTrip(t *testing.T) {
	enumOptions := &descriptorpb.EnumOptions{}
	proto.SetExtension(enumOptions, errors.E_DefaultStatus, int32(422))
	assertProtoRoundTrip(t, enumOptions, &descriptorpb.EnumOptions{})

	valueOptions := &descriptorpb.EnumValueOptions{}
	proto.SetExtension(valueOptions, errors.E_Options, &errors.Error{
		Status:  404,
		Reason:  "USER_NOT_FOUND",
		Message: "user not found",
	})
	assertProtoRoundTrip(t, valueOptions, &descriptorpb.EnumValueOptions{})
}

func assertProtoRoundTrip(t *testing.T, input, output proto.Message) {
	t.Helper()

	data, err := proto.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := proto.Unmarshal(data, output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(output, input) {
		t.Errorf("round trip mismatch:\n got: %v\nwant: %v", output, input)
	}
}
