package collection

import (
	"time"
)

const (
	// Bool : true or false
	Bool FieldType = "bool"
	// UInt8 :  0 to 255
	UInt8 FieldType = "uint8"
	// UInt16 :  0 to 65535
	UInt16 FieldType = "uint16"
	// UInt32 : 0 to 4294967295
	UInt32 FieldType = "uint32"
	// UInt64 : 0 to 18446744073709551615
	UInt64 FieldType = "uint64"
	// Int8 : -128 to 127
	Int8 FieldType = "int8"
	// Int16 : -32768 to 32767
	Int16 FieldType = "int16"
	// Int32 : -2147483648 to 2147483647
	Int32 FieldType = "int32"
	// Int64 : -9223372036854775808 to 9223372036854775807
	Int64 FieldType = "int64"
	// Float32 : -3.40282346638528859811704183484516925440e+38 to 3.40282346638528859811704183484516925440e+38
	Float32 FieldType = "float32"
	// Float64 : -1.797693134862315708145274237317043567981e+308 to 1.797693134862315708145274237317043567981e+308
	Float64 FieldType = "float64"
	// String : sequence of characters
	String FieldType = "string"
	// DateTime : '1970-01-01 00:00:01.000000000' UTC to '2038-01-19 03:14:07.999999999' UTC
	DateTime FieldType = "datetime"
	// Object : inner document
	Object FieldType = "object"
)

var (
	// FieldTypes - set of supported FieldType
	FieldTypes = map[FieldType]struct{}{
		Bool:     struct{}{},
		UInt8:    struct{}{},
		UInt16:   struct{}{},
		UInt32:   struct{}{},
		UInt64:   struct{}{},
		Int8:     struct{}{},
		Int16:    struct{}{},
		Int32:    struct{}{},
		Int64:    struct{}{},
		Float32:  struct{}{},
		Float64:  struct{}{},
		String:   struct{}{},
		DateTime: struct{}{},
		Object:   struct{}{},
	}

	dateFormats = map[string]struct{}{
		time.ANSIC:       struct{}{},
		time.UnixDate:    struct{}{},
		time.RubyDate:    struct{}{},
		time.RFC822:      struct{}{},
		time.RFC822Z:     struct{}{},
		time.RFC850:      struct{}{},
		time.RFC1123:     struct{}{},
		time.RFC1123Z:    struct{}{},
		time.RFC3339:     struct{}{},
		time.RFC3339Nano: struct{}{},
		time.Kitchen:     struct{}{},
		time.Stamp:       struct{}{},
		time.StampMilli:  struct{}{},
		time.StampMicro:  struct{}{},
		time.StampNano:   struct{}{},
	}
)
