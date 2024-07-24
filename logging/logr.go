package logging

import (
	"fmt"

	"go.opentelemetry.io/otel/log"
)

func kv2akv(keysAndValues ...any) []log.KeyValue {
	values := make([]log.KeyValue, 0, len(keysAndValues)/2)

	for i := 0; i+1 < len(keysAndValues); i += 2 {
		values = append(values, logKeyValueFromAny(
			keysAndValues[i],
			keysAndValues[i+1],
		))
	}

	return values
}

func logKeyValueFromAny(key, value any) log.KeyValue {
	var logValue log.Value

	switch typedValue := value.(type) {
	case bool:
		logValue = log.BoolValue(typedValue)
	case int:
		logValue = log.IntValue(typedValue)
	case int64:
		logValue = log.Int64Value(typedValue)
	case float64:
		logValue = log.Float64Value(typedValue)
	case []byte:
		logValue = log.BytesValue(typedValue)
	case []log.Value:
		logValue = log.SliceValue(typedValue...)
	case []log.KeyValue:
		logValue = log.MapValue(typedValue...)
	default:
		logValue = log.StringValue(fmt.Sprint(typedValue))
	}

	return log.KeyValue{
		Key:   fmt.Sprint(key),
		Value: logValue,
	}
}
