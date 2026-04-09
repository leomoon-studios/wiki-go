#!/usr/bin/env python3
"""
notion-to-wikigo.py — Import a Notion markdown export into a Wiki-Go document tree.

Usage:
    python3 notion-to-wikigo.py <notion-export.zip> <target-dir>
    python3 notion-to-wikigo.py <notion-export.zip> <target-dir> --dry-run
    python3 notion-to-wikigo.py <notion-export.zip> <target-dir> --force

Each Notion page becomes:
    <target-dir>/<slug-path>/document.md

Attachments (images, PDFs, …) are placed next to the document:
    <target-dir>/<slug-path>/<original-filename>

See README.md for how to export from Notion.
"""

import argparse
import os
import re
import shutil
import sys
import tempfile
import unicodedata
import zipfile
from pathlib import Path
from urllib.parse import unquote


# ---------------------------------------------------------------------------
# Slug helpers
# ---------------------------------------------------------------------------

# Notion appends a 32-char hex ID (and optional _all suffix) to every file/dir
_NOTION_ID_RE = re.compile(r'\s+[0-9a-f]{32}(_all)?$', re.IGNORECASE)


def strip_notion_id(name: str) -> str:
    """
    Remove the Notion page ID from the end of a file or directory name.
    The optional _all suffix (used in database CSV exports) is preserved.

        'My Page abc123def…'       →  'My Page'
        'Database abc123def…_all'  →  'Database_all'
    """
    def _replace(m: re.Match) -> str:
        return m.group(1) or ''   # keep '_all' if present, else remove everything
    return _NOTION_ID_RE.sub(_replace, name)


def to_slug(name: str) -> str:
    """
    Convert a name to a URL-friendly slug.

    'Engineering / Back-end'  →  'engineering-back-end'
    'Meeting Notes'           →  'meeting-notes'
    """
    # Decompose unicode characters so accents become separate combining marks
    name = unicodedata.normalize('NFD', name)
    # Drop the combining (accent) characters
    name = ''.join(c for c in name if unicodedata.category(c) != 'Mn')
    name = name.lower()
    # Replace anything that isn't alphanumeric with a hyphen
    name = re.sub(r'[^a-z0-9]+', '-', name)
    return name.strip('-') or 'untitled'


def path_to_slug_parts(parts: list[str]) -> list[str]:
    """Convert a list of path parts to slug parts, stripping Notion IDs."""
    result = []
    for part in parts:
        clean = strip_notion_id(part)
        if clean:
            result.append(to_slug(clean))
    return result


# ---------------------------------------------------------------------------
# Target path computation
# ---------------------------------------------------------------------------

def target_for_document(src_rel: Path, target_base: Path) -> Path:
    """
    Compute the Wiki-Go target path for a Notion .md file.

    Examples (relative to the extracted zip root):
        Areas/Business/Ideas.md  →  <target>/areas/business/ideas/document.md
        Home.md                  →  <target>/home/document.md
    """
    parts = list(src_rel.parts)

    # Strip .md and Notion ID from the filename
    stem = strip_notion_id(parts[-1].removesuffix('.md'))
    dir_parts = parts[:-1]

    slug_parts = path_to_slug_parts(dir_parts) + [to_slug(stem)]
    return target_base.joinpath(*slug_parts) / 'document.md'


def target_for_attachment(src_rel: Path, target_base: Path) -> Path:
    """
    Compute the Wiki-Go target path for a non-.md attachment.

    Notion places a page's attachments inside a same-named subdirectory.
    After import the attachment lives next to its document.md, so the
    directory part maps to the page slug and the filename is kept as-is
    (except that any Notion ID is stripped from the stem).

    Example:
        Areas/Business/image.png                →  <target>/areas/business/image.png
        Areas/Business/Data abc123…_all.csv     →  <target>/areas/business/Data_all.csv
    """
    parts = list(src_rel.parts)
    raw_filename = parts[-1]
    dir_parts = parts[:-1]

    # Strip Notion ID from the filename stem, keep extension
    p = Path(raw_filename)
    clean_stem = strip_notion_id(p.stem)
    filename = clean_stem + p.suffix if clean_stem else raw_filename

    slug_parts = path_to_slug_parts(dir_parts)
    return target_base.joinpath(*slug_parts) / filename


# ---------------------------------------------------------------------------
# Markdown content cleaning
# ---------------------------------------------------------------------------

# Attachment extensions whose link paths we rewrite to bare filenames
_ATTACHMENT_EXTS = {
    '.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.avif',
    '.pdf', '.csv', '.zip', '.drawio', '.reg',
    '.mp4', '.mov', '.avi', '.mkv',
    '.txt', '.log',
}

# Matches any markdown link or image: ![alt](url) or [text](url)
_LINK_RE = re.compile(r'(!?\[[^\]]*\]\()([^)\n]+)(\))')


def _fix_link(url: str) -> str:
    """
    If url is a local attachment path with a directory prefix, strip the
    prefix and return just the filename.

    'Folder/image.png'          →  'image.png'
    'A%20B/photo.jpg'           →  'photo.jpg'
    'https://example.com/x.png' →  unchanged
    'SubPage.md'                →  unchanged  (wiki internal link)
    """
    if url.startswith(('http://', 'https://', '#', '//')):
        return url
    if '/' not in url:
        return url

    decoded = unquote(url)
    filename = decoded.rsplit('/', 1)[-1]
    ext = Path(filename).suffix.lower()

    if ext not in _ATTACHMENT_EXTS:
        return url   # internal .md wiki link — leave alone

    return filename


def clean_content(content: str) -> str:
    """
    Clean up Notion-specific artefacts in markdown content:
    - Strip Notion page IDs from internal links
    - Rewrite attachment paths to bare filenames
    """
    # 1. Remove Notion IDs from internal .md links
    #    [Title](Path%20to%20Page%20HEXID.md)  →  [Title](Path%20to%20Page.md)
    content = re.sub(
        r'(\[([^\]]+)\]\([^)]*?)%20[0-9a-f]{32}(_all)?(\.md[^)]*\))',
        r'\1\4',
        content,
        flags=re.IGNORECASE,
    )

    # 2. Rewrite attachment paths to bare filenames
    def replacer(m: re.Match) -> str:
        prefix, url, suffix = m.group(1), m.group(2), m.group(3)
        return prefix + _fix_link(url) + suffix

    content = _LINK_RE.sub(replacer, content)
    return content


# ---------------------------------------------------------------------------
# Zip extraction
# ---------------------------------------------------------------------------

def extract_notion_zip(zip_path: Path, work_dir: Path) -> Path:
    """
    Extract a Notion export zip into work_dir.

    Notion sometimes wraps the real export in an outer zip that contains a
    single inner zip (e.g. ExportBlock-…-Part-1.zip). This function handles
    both single-level and double-level zips and returns the directory that
    directly contains the exported pages.
    """
    with zipfile.ZipFile(zip_path, 'r') as zf:
        # Check if the zip contains only one entry that is itself a zip
        entries = [e for e in zf.namelist() if not e.endswith('/')]
        if len(entries) == 1 and entries[0].lower().endswith('.zip'):
            # Outer wrapper — extract inner zip, then extract that
            inner_zip_path = work_dir / entries[0]
            zf.extractall(work_dir)
            inner_dir = work_dir / 'inner'
            inner_dir.mkdir()
            with zipfile.ZipFile(inner_zip_path, 'r') as inner_zf:
                inner_zf.extractall(inner_dir)
            return inner_dir
        else:
            zf.extractall(work_dir)
            return work_dir


# ---------------------------------------------------------------------------
# Main import logic
# ---------------------------------------------------------------------------

def import_notion(zip_path: Path, target_base: Path,
                  dry_run: bool = False, force: bool = False) -> None:
    stats = {'documents': 0, 'attachments': 0, 'skipped': 0, 'errors': 0}

    with tempfile.TemporaryDirectory(prefix='notion_import_') as tmp:
        tmp_path = Path(tmp)
        print(f'Extracting {zip_path.name} …')
        src_root = extract_notion_zip(zip_path, tmp_path)

        for root, dirs, files in os.walk(src_root):
            dirs.sort()
            root_path = Path(root)
            src_rel_dir = root_path.relative_to(src_root)

            for filename in sorted(files):
                src_file = root_path / filename
                src_rel = src_rel_dir / filename

                try:
                    if filename.lower().endswith('.md'):
                        target = target_for_document(src_rel, target_base)
                    else:
                        target = target_for_attachment(src_rel, target_base)

                    rel_target = target.relative_to(target_base)

                    if target.exists() and not force:
                        print(f'  [skip] {rel_target}')
                        stats['skipped'] += 1
                        continue

                    if filename.lower().endswith('.md'):
                        print(f'  [doc]  {src_rel}  →  {rel_target}')
                        if not dry_run:
                            target.parent.mkdir(parents=True, exist_ok=True)
                            text = src_file.read_text(encoding='utf-8', errors='replace')
                            target.write_text(clean_content(text), encoding='utf-8')
                        stats['documents'] += 1
                    else:
                        print(f'  [att]  {src_rel}  →  {rel_target}')
                        if not dry_run:
                            target.parent.mkdir(parents=True, exist_ok=True)
                            shutil.copy2(src_file, target)
                        stats['attachments'] += 1

                except Exception as exc:
                    print(f'  [err]  {src_rel}: {exc}', file=sys.stderr)
                    stats['errors'] += 1

    print()
    print('=' * 60)
    if dry_run:
        print('DRY RUN — no files were written.')
    print(
        f"Done: {stats['documents']} documents, "
        f"{stats['attachments']} attachments imported, "
        f"{stats['skipped']} skipped, "
        f"{stats['errors']} errors."
    )
    if stats['errors']:
        sys.exit(1)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description='Import a Notion markdown export into a Wiki-Go document tree.',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument('zip_file', type=Path,
                        help='Path to the Notion export .zip file')
    parser.add_argument('target_dir', type=Path,
                        help='Target directory (Wiki-Go docs/ folder)')
    parser.add_argument('--dry-run', action='store_true',
                        help='Show what would be done without writing any files')
    parser.add_argument('--force', action='store_true',
                        help='Overwrite existing files')
    args = parser.parse_args()

    if not args.zip_file.is_file():
        print(f'Error: {args.zip_file} is not a file.', file=sys.stderr)
        sys.exit(1)

    if not args.dry_run:
        args.target_dir.mkdir(parents=True, exist_ok=True)

    import_notion(
        zip_path=args.zip_file,
        target_base=args.target_dir,
        dry_run=args.dry_run,
        force=args.force,
    )


if __name__ == '__main__':
    main()
