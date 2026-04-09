# notion-to-wikigo

A Python script that imports a [Notion](https://notion.so) markdown export into a
[LeoMoon Wiki-Go](https://github.com/LeoMoonSecurity/wiki-go) document tree.

## What it does

- Converts each Notion page into `<slug-path>/document.md`
- Copies attachments (images, PDFs, …) next to their document
- Strips Notion's internal page IDs from filenames and links
- Fixes image/attachment paths so they resolve correctly in Wiki-Go
- Handles both flat exports and Notion's nested-zip format (large workspaces)

## Requirements

- Python 3.9+
- No external dependencies — standard library only

Install (nothing to install beyond Python itself):

```bash
pip install -r requirements.txt   # no-op, listed for completeness
```

## Exporting from Notion

1. Open Notion and go to **Settings → Export all workspace content**
   (or open a specific page and use **⋯ → Export**).

2. In the export dialog, configure:

   | Setting | Value |
   |---|---|
   | Export format | **Markdown & CSV** |
   | Include subpages | ✅ on |
   | **Create folders for subpages** | ✅ **on** ← required |
   | Include content | Everything |

   > **"Create folders for subpages" must be enabled.**
   > Without it Notion exports a flat list of files with no hierarchy,
   > and the script cannot reconstruct the page tree.

3. Click **Export**. Notion will email you a download link or offer a direct download.

4. Save the `.zip` file — you do not need to unzip it manually.

## Usage

```bash
python3 notion-to-wikigo.py <notion-export.zip> <target-dir> [options]
```

### Options

| Flag | Description |
|---|---|
| `--dry-run` | Show what would be imported without writing any files |
| `--force` | Overwrite existing files (default: skip) |

### Examples

Preview what will be imported:
```bash
python3 notion-to-wikigo.py export.zip ./docs --dry-run
```

Import into Wiki-Go's docs directory:
```bash
python3 notion-to-wikigo.py export.zip /path/to/wiki/docs
```

Import a second export without overwriting existing pages:
```bash
python3 notion-to-wikigo.py export2.zip /path/to/wiki/docs
```

Re-import and overwrite everything:
```bash
python3 notion-to-wikigo.py export.zip /path/to/wiki/docs --force
```

## Output structure

Given this Notion page hierarchy:

```
Engineering/
  Back-end/
    API Design.md
    API Design/
      diagram.png
    Database Schema.md
  Onboarding.md
Handbook.md
```

the script produces:

```
docs/
  engineering/
    back-end/
      api-design/
        document.md      ← page content
        diagram.png      ← attachment, placed next to its document
      database-schema/
        document.md
    onboarding/
      document.md
  handbook/
    document.md
```

Page names are converted to lowercase slugs with accented characters
transliterated to ASCII (`Björn Åkesson` → `bjorn-akesson`,
`Ré́union` → `reunion`).

## Large workspaces — split exports

For large workspaces Notion may split the export into multiple zip files
(`Part-1.zip`, `Part-2.zip`, …). Run the script once per file, pointing at
the same target directory:

```bash
python3 notion-to-wikigo.py export-part1.zip ./docs
python3 notion-to-wikigo.py export-part2.zip ./docs
```

Already-imported pages are skipped automatically. Use `--force` to re-import.

## Known limitations

- **Internal wiki links** — Notion exports sub-page links as relative file paths
  (e.g. `[Schema](Database%20Schema.md)`). These are not automatically converted
  to Wiki-Go URLs and may need manual editing after import.
- **Databases** — Notion database views are exported as CSV files. The script
  copies them as attachments but does not convert them into markdown tables.
- **Synced blocks** — Content from synced blocks may appear duplicated in the
  export; this is a Notion limitation outside the script's control.
