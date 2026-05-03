# BACKLOG — Post-MVP / Future Work

> Ticket di sini **tidak** ada di scope v0.5. Akan ditinjau setelah MVP solid.

## Daftar Ticket

| ID | Title | Type | Points | Priority |
|---|---|---|---|---|
| [OFFICE-901](#office-901) | Encryption / OOXML decryption support | Story | 21 | P3 |
| [OFFICE-902](#office-902) | Chart structure (read DrawingML chart XML) | Story | 13 | P3 |
| [OFFICE-903](#office-903) | Pivot table read-only | Story | 13 | P3 |
| [OFFICE-904](#office-904) | DOCX header/footer write support | Story | 8 | P3 |

---

## OFFICE-901

### [Story] Encryption / OOXML decryption support

```
Type     : Story
Priority : P3
Points   : 21
Sprint   : Backlog (post-v1.0)
Epic     : Future
```

> Decrypt password-protected OOXML via MS-OFFCRYPTO. Stretch beyond MVP.
>
> Spec: [MS-OFFCRYPTO](https://learn.microsoft.com/en-us/openspecs/office_file_formats/ms-offcrypto/) — Office Document Cryptography Structure.
> Implementasi membutuhkan: AES-128/256, SHA-1/512, ECMA-376 Agile Encryption, password derivation.
> **Risiko:** Crypto implementation sensitive — wajib pakai stdlib crypto, audit menyeluruh.

---

## OFFICE-902

### [Story] Chart structure (read DrawingML chart XML)

```
Type     : Story
Priority : P3
Points   : 13
Sprint   : Backlog (Tier 3, post-MVP)
```

> Parse `chart1.xml` ke struct dengan series + axis. Tidak rendering, hanya data.
>
> Scope: column, bar, line, pie, scatter. Series, categories, values, axis labels.
> Out of scope: combo chart, 3D, sparkline.

---

## OFFICE-903

### [Story] Pivot table read-only

```
Type     : Story
Priority : P3
Points   : 13
Sprint   : Backlog
```

> Parse `pivotTable*.xml` + `pivotCacheDefinition*.xml` + `pivotCacheRecords*.xml` untuk akses ke struktur pivot.
>
> Scope: row/column/data fields, filters, cache records iterator.
> Out of scope: re-compute pivot, slicers, timeline.

---

## OFFICE-904

### [Story] DOCX header/footer write support

```
Type     : Story
Priority : P3
Points   : 8
Sprint   : Backlog (Tier 2)
```

> Add API untuk membuat/mengedit `header*.xml` dan `footer*.xml`, termasuk per-section override (first/odd/even).
>
> Scope: text, page number, date field.
> Dependency: OFFICE-107 (DOCX writer foundation).

---

**Navigasi:** [← E08 Security](./E08-security.md) | [Index](./README.md) | [Definition of Done →](./definition-of-done.md)
