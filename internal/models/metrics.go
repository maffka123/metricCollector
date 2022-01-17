package models

type Metrics struct {
	ID    string   `json:"id"`              // metrics name
	MType string   `json:"type"`            // metrics type: can be counter or gauge
	Delta *int64   `json:"delta,omitempty"` // metrics values if type is counter
	Value *float64 `json:"value,omitempty"` // metrics values if type is gauge
}
