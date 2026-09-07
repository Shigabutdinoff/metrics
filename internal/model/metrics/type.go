package metrics

// Поддерживаемые типы метрик.
const (
	// Counter метрика-счётчик: значения накапливаются суммированием.
	Counter Type = "counter"
	// Gauge метрика-измерение: новое значение вытесняет прежнее.
	Gauge Type = "gauge"
)
