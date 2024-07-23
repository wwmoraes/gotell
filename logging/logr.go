package logging

import (
	"fmt"

	"go.opentelemetry.io/otel/log"
)

func kv2akv(keysAndValues ...any) []log.KeyValue {
	values := make([]log.KeyValue, 0, len(keysAndValues)/2)

	for i := 0; i+1 < len(keysAndValues); i += 2 {
		values = append(values, log.String(
			fmt.Sprint(keysAndValues[i]),
			fmt.Sprint(keysAndValues[i+1]),
		))
	}

	return values
}
