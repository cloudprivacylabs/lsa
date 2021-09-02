package gl

// NullValue is the null script result value
type NullValue struct {
	basicValue
}

// AsString returns "null"
func (NullValue) AsString() (string, error) { return "null", nil }

// Eq returns true only if the value is a null value
func (NullValue) Eq(v Value) (bool, error) {
	if v == nil {
		return true, nil
	}
	if _, ok := v.(NullValue); ok {
		return true, nil
	}
	return false, nil
}
