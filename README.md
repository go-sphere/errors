# errors proto

This is a proto file that defines the sphere errors framework with status codes and messages. It provides extensions for protobuf enums to support structured error handling in Go applications.

## Features

- Enum-level default status code configuration
- Per-value error customization (status, reason, message)
- Integration with HTTP status codes
- Support for custom error messages and reasons
- Designed for use with `protoc-gen-sphere-errors` plugin

## Proto Definition

The core proto file defines:

```protobuf
syntax = "proto3";

package sphere.errors;

import "google/protobuf/descriptor.proto";

message Error {
  int32 status = 1;    // HTTP status code
  string reason = 2;   // Optional reason string
  string message = 3;  // Human-readable error message
}

extend google.protobuf.EnumOptions {
  int32 default_status = 18534200;  // Default status for entire enum
}

extend google.protobuf.EnumValueOptions {
  Error options = 18534210;  // Per-value error configuration
}
```

## Usage Example

### Basic Error Enum

```protobuf
syntax = "proto3";

package api.v1;

import "sphere/errors/errors.proto";

enum UserError {
  option (sphere.errors.default_status) = 500;  // Default status for all values
  
  USER_ERROR_UNSPECIFIED = 0;
  USER_ERROR_NOT_FOUND = 1001 [(sphere.errors.options) = {
    status: 404
    message: "用户不存在"
  }];
  USER_ERROR_INVALID_EMAIL = 1002 [(sphere.errors.options) = {
    status: 400
    message: "邮箱格式不正确"
  }];
  USER_ERROR_PERMISSION_DENIED = 1003 [(sphere.errors.options) = {
    status: 403
    message: "权限不足"
  }];
}
```

### Advanced Example with Reasons

```protobuf
enum PaymentError {
  option (sphere.errors.default_status) = 500;
  
  PAYMENT_ERROR_UNSPECIFIED = 0;
  PAYMENT_ERROR_INSUFFICIENT_FUNDS = 2001 [(sphere.errors.options) = {
    status: 400
    reason: "insufficient_funds"
    message: "账户余额不足"
  }];
  PAYMENT_ERROR_CARD_EXPIRED = 2002 [(sphere.errors.options) = {
    status: 400
    reason: "card_expired"
    message: "信用卡已过期"
  }];
  PAYMENT_ERROR_PROCESSING_FAILED = 2003 [(sphere.errors.options) = {
    status: 502
    reason: "processing_failed"
    message: "支付处理失败，请稍后重试"
  }];
}
```

## Integration with buf

Add this dependency to your `buf.yaml`:

```yaml
version: v2
deps:
  - buf.build/go-sphere/errors
```

Configure code generation in `buf.gen.yaml`:

```yaml
version: v2
managed:
  enabled: true
  disable:
    - file_option: go_package_prefix
      module: buf.build/go-sphere/errors
plugins:
  - local: protoc-gen-sphere-errors
    out: api
    opt: paths=source_relative
```

## Generated Code Usage

When used with `protoc-gen-sphere-errors`, the following Go code is generated:

```go
// Direct error return
return nil, apiv1.UserError_USER_ERROR_NOT_FOUND

// With error wrapping
if err != nil {
    return nil, apiv1.UserError_USER_ERROR_NOT_FOUND.Join(err)
}

// With custom message
return nil, apiv1.UserError_USER_ERROR_NOT_FOUND.JoinWithMessage("User ID: 123", err)
```

## Error Configuration Options

### Enum Level Options

- `default_status`: Sets the default HTTP status code for all enum values that don't specify their own status

### Enum Value Options

- `status`: HTTP status code (overrides default_status)
- `reason`: Optional machine-readable reason code
- `message`: Human-readable error message for client display

## Best Practices

1. **Use meaningful error codes**: Choose enum values that clearly indicate the error type
2. **Set appropriate HTTP status codes**: Use standard HTTP status codes (400, 401, 403, 404, 500, etc.)
3. **Provide clear messages**: Write user-friendly error messages in the appropriate language
4. **Use reasons for API consumers**: Include reason strings for programmatic error handling
5. **Group related errors**: Keep related errors in the same enum for better organization

## Common HTTP Status Codes

- `400`: Bad Request - Client error, invalid input
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Permission denied
- `404`: Not Found - Resource doesn't exist
- `409`: Conflict - Resource conflict
- `422`: Unprocessable Entity - Validation failed
- `429`: Too Many Requests - Rate limiting
- `500`: Internal Server Error - Server-side error
- `502`: Bad Gateway - External service error
- `503`: Service Unavailable - Service temporarily down