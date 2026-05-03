package opcprops

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const sampleCore = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties
  xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Test Document</dc:title>
  <dc:subject>Testing</dc:subject>
  <dc:creator>Alice</dc:creator>
  <cp:lastModifiedBy>Bob</cp:lastModifiedBy>
  <cp:revision>3</cp:revision>
  <dcterms:created xsi:type="dcterms:W3CDTF">2024-01-15T10:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2024-06-20T12:30:00Z</dcterms:modified>
</cp:coreProperties>`

const sampleApp = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>Microsoft Office Word</Application>
  <AppVersion>16.0000</AppVersion>
  <Company>Contoso</Company>
  <DocSecurity>0</DocSecurity>
  <ScaleCrop>false</ScaleCrop>
  <LinksUpToDate>false</LinksUpToDate>
  <SharedDoc>false</SharedDoc>
  <HyperlinksChanged>false</HyperlinksChanged>
</Properties>`

func TestParseCoreFields(t *testing.T) {
	c, err := ParseCore(strings.NewReader(sampleCore))
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Test Document" {
		t.Errorf("Title: got %q", c.Title)
	}
	if c.Creator != "Alice" {
		t.Errorf("Creator: got %q", c.Creator)
	}
	if c.LastModifiedBy != "Bob" {
		t.Errorf("LastModifiedBy: got %q", c.LastModifiedBy)
	}
	if c.Revision != "3" {
		t.Errorf("Revision: got %q", c.Revision)
	}
	if c.Created.IsZero() {
		t.Error("Created should not be zero")
	}
	if c.Modified.IsZero() {
		t.Error("Modified should not be zero")
	}
}

func TestParseAppFields(t *testing.T) {
	a, err := ParseApp(strings.NewReader(sampleApp))
	if err != nil {
		t.Fatal(err)
	}
	if a.Application != "Microsoft Office Word" {
		t.Errorf("Application: got %q", a.Application)
	}
	if a.AppVersion != "16.0000" {
		t.Errorf("AppVersion: got %q", a.AppVersion)
	}
	if a.Company != "Contoso" {
		t.Errorf("Company: got %q", a.Company)
	}
}

func TestParseCoreRoundTrip(t *testing.T) {
	c, err := ParseCore(strings.NewReader(sampleCore))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	c2, err := ParseCore(&buf)
	if err != nil {
		t.Fatalf("parse after WriteTo: %v", err)
	}

	if c2.Title != c.Title {
		t.Errorf("Title: got %q, want %q", c2.Title, c.Title)
	}
	if c2.Creator != c.Creator {
		t.Errorf("Creator: got %q, want %q", c2.Creator, c.Creator)
	}
	if c2.LastModifiedBy != c.LastModifiedBy {
		t.Errorf("LastModifiedBy: got %q, want %q", c2.LastModifiedBy, c.LastModifiedBy)
	}
	if c2.Revision != c.Revision {
		t.Errorf("Revision: got %q, want %q", c2.Revision, c.Revision)
	}
	if !c2.Created.Equal(c.Created) {
		t.Errorf("Created: got %v, want %v", c2.Created, c.Created)
	}
}

func TestParseAppRoundTrip(t *testing.T) {
	a, err := ParseApp(strings.NewReader(sampleApp))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := a.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	a2, err := ParseApp(&buf)
	if err != nil {
		t.Fatalf("parse after WriteTo: %v", err)
	}

	if a2.Application != a.Application {
		t.Errorf("Application: got %q, want %q", a2.Application, a.Application)
	}
	if a2.Company != a.Company {
		t.Errorf("Company: got %q, want %q", a2.Company, a.Company)
	}
}

func TestDefaultApplication(t *testing.T) {
	a := &AppProperties{}
	var buf bytes.Buffer
	if _, err := a.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	a2, err := ParseApp(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if a2.Application != Application {
		t.Errorf("default Application: got %q, want %q", a2.Application, Application)
	}
	if a2.AppVersion != Version {
		t.Errorf("default AppVersion: got %q, want %q", a2.AppVersion, Version)
	}
}

func TestCoreTimestamps(t *testing.T) {
	ts := time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC)
	c := &CoreProperties{
		Title:   "TS Test",
		Created: ts,
	}
	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "2024-03-15T09:30:00Z") {
		t.Errorf("expected RFC3339 timestamp in output, got: %s", s)
	}
}

func TestCoreWriteToXMLDecl(t *testing.T) {
	c := &CoreProperties{Title: "X"}
	var buf bytes.Buffer
	c.WriteTo(&buf)
	if !strings.HasPrefix(buf.String(), `<?xml`) {
		t.Error("WriteTo should start with XML declaration")
	}
}

func TestAppWriteToXMLDecl(t *testing.T) {
	a := &AppProperties{}
	var buf bytes.Buffer
	a.WriteTo(&buf)
	if !strings.HasPrefix(buf.String(), `<?xml`) {
		t.Error("WriteTo should start with XML declaration")
	}
}

func TestMalformedCore(t *testing.T) {
	// Malformed XML — should return error (not panic)
	_, err := ParseCore(strings.NewReader("<not closed"))
	if err == nil {
		t.Error("expected error parsing malformed XML")
	}
}

func TestMalformedApp(t *testing.T) {
	_, err := ParseApp(strings.NewReader("not xml"))
	if err == nil {
		t.Error("expected error parsing malformed XML")
	}
}

func TestNilCoreWriteTo(t *testing.T) {
	var c *CoreProperties
	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Errorf("nil CoreProperties.WriteTo should not error: %v", err)
	}
}

func TestNilAppWriteTo(t *testing.T) {
	var a *AppProperties
	var buf bytes.Buffer
	if _, err := a.WriteTo(&buf); err != nil {
		t.Errorf("nil AppProperties.WriteTo should not error: %v", err)
	}
}
