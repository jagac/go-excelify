package types

type RequestJson struct {
	Filename string                   `json:"filename"`
	Data     []map[string]interface{} `json:"data"`
	Meta     struct {
		Columns []ColumnMeta `json:"columns"`
	} `json:"meta"`
}

type ColumnMeta struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	DefaultVisibility string `json:"default_visibility,omitempty"`
}
type MetaData struct {
	Columns []ColumnMeta `json:"columns"`
}