package cmderr

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		if details, ok := se.Details["resource"]; ok {
			return details != nil
		}
	}
	return false
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		if field, ok := se.Details["field"]; ok {
			return field != nil
		}
	}
	return false
}

// IsConnectionError checks if an error is a connection error.
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		return se.Operation == "connect to iTerm2"
	}
	return false
}
