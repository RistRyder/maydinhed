package stores

type Command[K StoreKey] struct {
	Key       K        `json:"key,omitempty"`
	Operation string   `json:"op,omitempty"`
	Value     Location `json:"value,omitempty"`
}
