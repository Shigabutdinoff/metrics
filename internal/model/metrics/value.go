package metrics

type (
	// Type тип метрики: Counter или Gauge.
	Type string
	// CounterValue значение counter-метрики, nil отличает 0 от пустого.
	CounterValue *int64
	// GaugeValue значение gauge-метрики, nil отличает 0 от пустого.
	GaugeValue *float64
)
