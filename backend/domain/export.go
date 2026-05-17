package domain

type ExportReport struct {
	Path           string `json:"path"`
	ItemsWritten   int    `json:"itemsWritten"`
	BytesWritten   int64  `json:"bytesWritten"`
	DurationMillis int64  `json:"durationMillis"`
}
