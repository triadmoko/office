package docx

// marshalWordSettingsXML returns minimal word/settings.xml for new packages.
// updateFields prompts Microsoft Word to refresh fields (e.g. TOC, PAGE) when the document is opened.
// Other office suites may ignore this flag.
func marshalWordSettingsXML() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:settings xmlns:w="` + nsW + `">` +
		`<w:updateFields w:val="true"/>` +
		`</w:settings>`)
}
