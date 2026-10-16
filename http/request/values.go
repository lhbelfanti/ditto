package request

import (
	"fmt"
	"time"
)

// String returns the String param name, or "" when an optional one was left out.
func (v Values) String(name string) string {
	return get[string](v, name, kindString)
}

// Int returns the Int param name, or 0 when an optional one was left out.
func (v Values) Int(name string) int {
	return get[int](v, name, kindInt)
}

// Bool returns the Bool param name, or false when an optional one was left out.
func (v Values) Bool(name string) bool {
	return get[bool](v, name, kindBool)
}

// Time returns the Time param name, or the zero time when an optional one was left out.
func (v Values) Time(name string) time.Time {
	return get[time.Time](v, name, kindTime)
}

// Has reports whether the request sent the param name, to tell an optional param left out from one sent with its zero value.
func (v Values) Has(name string) bool {
	_, sent := v.values[name]

	return sent
}

// get panics when name was not declared with the kind of the getter: that is a bug in the handler, not in the request.
func get[T any](v Values, name string, k kind) T {
	field, declared := v.fields[name]
	if !declared || field.kind != k {
		panic(fmt.Sprintf("request: %q is not declared as %s", name, k))
	}

	value, _ := v.values[name].(T)

	return value
}
