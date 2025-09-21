package proto

// BoolPointer returns a pointer to a bool value.
func BoolPointer(b bool) *bool {
	return &b
}

// StringPointer returns a pointer to a string value.
func StringPointer(s string) *string {
	return &s
}

// Int32Pointer returns a pointer to an int32 value.
func Int32Pointer(i int32) *int32 {
	return &i
}

// Float32Pointer returns a pointer to a float32 value.
func Float32Pointer(f float32) *float32 {
	return &f
}
