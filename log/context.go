package log

import (
	"context"
	"maps"
	"time"

	"github.com/rs/zerolog"
)

// Param creates a new field to be saved into context. It is logged with every call made with that
// context, and it is not embedded in errors; see EmbeddedParam.
func Param(key string, value any) field {
	return field{Key: key, Value: value}
}

// EmbeddedParam creates a field like Param that NewFieldsError also embeds in the text of the
// errors it builds. Use it for the fields a function adds with With and that explain a failure
// further up. Fields the logging point already adds itself, such as the request ID, stay Param
// so they are not repeated in the error.
func EmbeddedParam(key string, value any) field {
	return field{Key: key, Value: value, embed: true}
}

// With returns a context carrying fields on top of the ones ctx already carries. It copies the
// existing fields instead of writing into them, so ctx keeps its own and contexts derived from the
// same parent, even from different goroutines, never see each other's fields.
func With(ctx context.Context, fields ...field) context.Context {
	existing, _ := ctx.Value(logCtxKey{}).(map[string]any)
	params := make(map[string]any, len(existing)+len(fields))
	maps.Copy(params, existing)

	existingEmbedded, _ := ctx.Value(embedCtxKey{}).(map[string]bool)
	embedded := make(map[string]bool, len(existingEmbedded)+len(fields))
	maps.Copy(embedded, existingEmbedded)

	for _, t := range fields {
		params[t.Key] = t.Value
		embedded[t.Key] = t.embed
	}

	return context.WithValue(context.WithValue(ctx, logCtxKey{}, params), embedCtxKey{}, embedded)
}

// NewFieldsError returns err with the fields that ctx carries as EmbeddedParam embedded in its
// text, so they reach the place where the error is logged even though a context created below that
// place is not visible from it. Use it in the function that adds those fields with With, on the
// errors that function returns. It returns err unchanged when ctx carries no embedded field, and
// nil when err is nil.
func NewFieldsError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	fields, _ := ctx.Value(logCtxKey{}).(map[string]any)
	embedded, _ := ctx.Value(embedCtxKey{}).(map[string]bool)

	selected := make(map[string]any, len(embedded))
	for key, embed := range embedded {
		if embed {
			selected[key] = fields[key]
		}
	}
	if len(selected) == 0 {
		return err
	}

	return &FieldsError{cause: err, fields: selected}
}

func withContextParams(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if ctx == nil {
		return event
	}

	params, ok := ctx.Value(logCtxKey{}).(map[string]any)
	if !ok {
		return event
	}

	// If a specific type (supported by zerolog.Event) is needed, add it to the switch
	for key, value := range params {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case float64:
			event = event.Float64(key, v)
		case bool:
			event = event.Bool(key, v)
		case error:
			event = event.AnErr(key, v)
		case []string:
			event = event.Strs(key, v)
		case []int:
			event = event.Ints(key, v)
		case []float64:
			event = event.Floats64(key, v)
		case []byte:
			event = event.Bytes(key, v)
		case time.Time:
			event = event.Time(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	return event
}
