package request

import "time"

// String returns the String param name, or "" when it was not sent or not declared as String.
func (v Values) String(name string) string {
	return get[string](v, name)
}

// Int returns the Int param name, or 0 when it was not sent or not declared as Int.
func (v Values) Int(name string) int {
	return get[int](v, name)
}

// Bool returns the Bool param name, or false when it was not sent or not declared as Bool.
func (v Values) Bool(name string) bool {
	return get[bool](v, name)
}

// Time returns the Time param name, or the zero time when it was not sent or not declared as Time.
func (v Values) Time(name string) time.Time {
	return get[time.Time](v, name)
}

// Has reports whether the request sent the param name, to tell an optional param left out from one sent with its zero value.
func (v Values) Has(name string) bool {
	_, sent := v.values[name]

	return sent
}

func get[T any](v Values, name string) T {
	value, _ := v.values[name].(T)

	return value
}
