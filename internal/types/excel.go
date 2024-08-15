package types

type ExcelStyles struct {
	HeaderStyle     int
	IntStyle        int
	FloatStyle      int
	DatetimeStyle   int
	PercentageStyle int
	TextStyle       int
}

type CellData struct {
	RowIndex int
	ColIndex int
	Value    interface{}
	Style    int
}
