package wml

import (
	"encoding/xml"

	"github.com/triadmoko/office/internal/ooxml"
)

func isWML(space, local string) bool {
	if local == "" {
		return false
	}
	return space == "" || space == ooxml.NSWordprocessingML ||
		space == ooxml.NSWordprocessingML14 || space == ooxml.NSWordprocessingML15
}

const nsXML = "http://www.w3.org/XML/1998/namespace"

func isXMLSpace(name xml.Name) bool {
	return name.Local == "space" && (name.Space == nsXML || name.Space == "")
}
