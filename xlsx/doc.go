// Package xlsx reads and writes SpreadsheetML (.xlsx) workbooks.
//
// Read: [Open], [Workbook.Sheets], [Sheet.Rows] streaming, [Sheet.Cell] random access,
// shared strings, styles subset, layout (merge, hidden, freeze), and [Workbook.SharedString].
//
// Write: [NewWorkbook], [Workbook.AddSheet], [Sheet.SetCell], [Sheet.SetFormula], [Sheet.SetHyperlink],
// [Sheet.StreamWriter] for large exports, and [Workbook.Save] / [Workbook.SaveFile].
// Opening a file snapshots all OPC parts so [Save] on an opened workbook round-trips the package.
package xlsx
