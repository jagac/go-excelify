package converter

import (
	"github.com/xuri/excelize/v2"
)

type ExcelStyles struct {
	HeaderStyle     int
	IntStyle        int
	FloatStyle      int
	DatetimeStyle   int
	PercentageStyle int
	TextStyle       int
	HiddenStyle     int
}

func createStyles(f *excelize.File) (*ExcelStyles, error) {
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Aptos Narrow",
			Bold:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	intStyle, err := f.NewStyle(&excelize.Style{NumFmt: 1})
	if err != nil {
		return nil, err
	}

	floatStyle, err := f.NewStyle(&excelize.Style{NumFmt: 2})
	if err != nil {
		return nil, err
	}

	exp := "yyyy-mm-dd"
	datetimeStyle, err := f.NewStyle(&excelize.Style{CustomNumFmt: &exp})
	if err != nil {
		return nil, err
	}

	percentageStyle, err := f.NewStyle(&excelize.Style{NumFmt: 10})
	if err != nil {
		return nil, err
	}

	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}

	hiddenFontColorStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: "#ff00ff",
		},
	})
	if err != nil {
		return nil, err
	}

	return &ExcelStyles{
		HeaderStyle:     headerStyle,
		IntStyle:        intStyle,
		FloatStyle:      floatStyle,
		DatetimeStyle:   datetimeStyle,
		PercentageStyle: percentageStyle,
		TextStyle:       textStyle,
		HiddenStyle:     hiddenFontColorStyle,
	}, nil
}
