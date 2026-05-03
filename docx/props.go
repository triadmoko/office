package docx

import (
	"time"
)

const defaultCreator = "github.com/triadmoko/office"

func marshalCoreProps() []byte {
	created := time.Now().UTC().Format(time.RFC3339)
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" ` +
		`xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" ` +
		`xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:creator>` + escapeCharData(defaultCreator) + `</dc:creator>` +
		`<cp:lastModifiedBy>` + escapeCharData(defaultCreator) + `</cp:lastModifiedBy>` +
		`<dcterms:created xsi:type="dcterms:W3CDTF">` + escapeCharData(created) + `</dcterms:created>` +
		`<dcterms:modified xsi:type="dcterms:W3CDTF">` + escapeCharData(created) + `</dcterms:modified>` +
		`</cp:coreProperties>`)
}

func marshalAppProps() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" ` +
		`xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">` +
		`<Application>Microsoft Office Word</Application>` +
		`<DocSecurity>0</DocSecurity>` +
		`<ScaleCrop>false</ScaleCrop>` +
		`<Company></Company>` +
		`<LinksUpToDate>false</LinksUpToDate>` +
		`<SharedDoc>false</SharedDoc>` +
		`<HyperlinksChanged>false</HyperlinksChanged>` +
		`<AppVersion>16.0000</AppVersion>` +
		`</Properties>`)
}
