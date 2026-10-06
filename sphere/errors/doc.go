// Package errors holds the Go bindings for sphere/errors/errors.proto, the
// protobuf options that declare API error codes as enums, plus [NewError],
// the runtime constructor that the generated error helpers call.
//
// An enum becomes an error enum when it sets the (sphere.errors.default_status)
// enum option. Each value may then carry (sphere.errors.options) with an
// [Error] message giving its HTTP status, reason, and user-facing message.
// The protoc-gen-sphere-errors plugin reads these options and generates
// Error, GetCode, GetStatus, GetMessage, Join, and JoinWithMessage methods on
// the enum; Join and JoinWithMessage build their result with NewError.
//
// # Declaring errors in .proto
//
// Depend on buf.build/go-sphere/errors, then:
//
//	import "sphere/errors/errors.proto";
//
//	enum UserError {
//	  option (sphere.errors.default_status) = 500;
//	  USER_ERROR_UNSPECIFIED = 0;
//	  USER_ERROR_NOT_FOUND = 1001 [(sphere.errors.options) = {
//	    status: 404
//	    reason: "USER_NOT_FOUND"
//	    message: "user does not exist"
//	  }];
//	}
//
// # Go usage
//
// Application code normally uses the generated enum methods, for example
// UserError_USER_ERROR_NOT_FOUND.Join(err). Code generators and tools read the
// options from descriptors with [E_DefaultStatus] and [E_Options]:
//
//	import (
//		"github.com/go-sphere/errors/sphere/errors"
//		"google.golang.org/protobuf/proto"
//	)
//
//	opts := enumValueDescriptor.Options() // protoreflect.EnumValueDescriptor
//	if e, ok := proto.GetExtension(opts, errors.E_Options).(*errors.Error); ok && e != nil {
//		status, reason, message := e.GetStatus(), e.GetReason(), e.GetMessage()
//	}
//
// To build an HTTP-classifiable error by hand, call NewError directly:
//
//	err := errors.NewError(404, 1001, "user does not exist", cause)
//
// The result exposes GetStatus, GetCode, and GetMessage methods, which
// adapters such as github.com/go-sphere/httpx use to render responses.
// This package depends only on the standard library and protobuf.
package errors
