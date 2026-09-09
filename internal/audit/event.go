package audit

// Event событие аудита принятых сервером метрик.
type Event struct {
	// TS момент приёма метрик в секундах Unix.
	TS int64 `json:"ts"`
	// Metrics имена принятых метрик.
	Metrics []string `json:"metrics"`
	// IPAddress адрес клиента, приславшего метрики.
	IPAddress string `json:"ip_address"`
}
