package analysis

import (
	"bytes"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

func ExtractTextFromPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for pageIndex := 1; pageIndex <= reader.NumPage(); pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		builder.WriteString(text)
		builder.WriteByte('\n')
	}

	return NormalizeExtractedText(builder.String()), nil
}
