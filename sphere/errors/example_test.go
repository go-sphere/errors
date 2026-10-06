package errors_test

import (
	stderrors "errors"
	"fmt"
	"io/fs"

	"github.com/go-sphere/errors/sphere/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ExampleNewError() {
	err := errors.NewError(404, 1001, "user does not exist", fs.ErrNotExist)

	// Adapters classify the error through its getter methods rather than a
	// concrete type.
	var coded interface {
		GetStatus() int32
		GetCode() int32
		GetMessage() string
	}
	if stderrors.As(err, &coded) {
		fmt.Println(coded.GetStatus(), coded.GetCode(), coded.GetMessage())
	}
	fmt.Println(err)
	fmt.Println(stderrors.Is(err, fs.ErrNotExist))
	// Output:
	// 404 1001 user does not exist
	// file does not exist
	// true
}

func ExampleNewError_nilCause() {
	// With a nil cause, the error text comes from the HTTP status.
	fmt.Println(errors.NewError(404, 1001, "", nil))
	fmt.Println(errors.NewError(799, 1002, "", nil))
	// Output:
	// Not Found
	// Unknown error
}

func Example_enumValueOptions() {
	// A code generator receives these options from the enum value descriptor.
	// Here they are built by hand to show the round trip.
	opts := &descriptorpb.EnumValueOptions{}
	proto.SetExtension(opts, errors.E_Options, &errors.Error{
		Status:  404,
		Reason:  "USER_NOT_FOUND",
		Message: "user does not exist",
	})

	if !proto.HasExtension(opts, errors.E_Options) {
		fmt.Println("not an annotated value")
		return
	}
	e, ok := proto.GetExtension(opts, errors.E_Options).(*errors.Error)
	if !ok || e == nil {
		fmt.Println("unexpected extension type")
		return
	}
	fmt.Println(e.GetStatus(), e.GetReason(), e.GetMessage())
	// Output: 404 USER_NOT_FOUND user does not exist
}

func Example_enumDefaultStatus() {
	opts := &descriptorpb.EnumOptions{}
	proto.SetExtension(opts, errors.E_DefaultStatus, int32(500))

	// protoc-gen-sphere-errors only treats enums carrying this option as error enums.
	if proto.HasExtension(opts, errors.E_DefaultStatus) {
		status, _ := proto.GetExtension(opts, errors.E_DefaultStatus).(int32)
		fmt.Println(status)
	}
	// Output: 500
}
