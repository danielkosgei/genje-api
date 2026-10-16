"""Ingest politician photos from official sources, with the source recorded for each one.

Input: a JSON list, one entry per photo (see photos.example.json):
  slug             the politician's slug in the database
  photo_url        the image's address on the official site
  source_name      the official publisher, e.g. "IEBC" or the politician's own site
  source_page_url  the official page the photo was found on (the citation)
  source_type      "official" for government, party or politician's own publications

For each entry the script downloads the image, checks that the bytes really are a JPEG or PNG, saves it to
static/photos/<slug>.<ext>, and records the SHA-256, the byte count, the source and the time retrieved. The politician's
photo_url is set to the stored file only after the file is saved, and the whole database change runs in one transaction.

Usage (from the repo root):
  python tools/ingest/photos.py tools/ingest/photos.json
Requires DATABASE_URL in the environment, and psql on the PATH.
"""
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
PHOTO_DIR = os.path.join(ROOT, "static", "photos")
LOG = os.path.join(ROOT, "tools", "ingest", "photos_log.json")
USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124.0 Safari/537.36 JaladaIngestion/1.0"
)
MAGIC = {b"\xff\xd8\xff": ("jpg", "image/jpeg"), b"\x89PNG\r\n\x1a\n": ("png", "image/png")}
SLUG = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")
MAX_BYTES = 5 * 1024 * 1024


def lit(value):
    """A SQL string literal. Only used for values this script has already validated."""
    return "'" + str(value).replace("'", "''") + "'"


def download(url):
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT, "Accept": "image/*"})
    with urllib.request.urlopen(req, timeout=30) as r:
        data = r.read(MAX_BYTES + 1)
    if len(data) > MAX_BYTES:
        raise ValueError("larger than 5 MB")
    return data


def image_kind(data):
    for magic, kind in MAGIC.items():
        if data.startswith(magic):
            return kind
    return None


def record_sql(entry, data, ext, content_type, digest):
    slug = entry["slug"]
    file_path = f"static/photos/{slug}.{ext}"
    now = datetime.now(timezone.utc).isoformat()
    return f"""
INSERT INTO sources (name, url, type, reliability, last_accessed_at)
SELECT {lit(entry["source_name"])}, {lit(entry["source_page_url"])}, {lit(entry["source_type"])}, 'official', {lit(now)}
WHERE NOT EXISTS (SELECT 1 FROM sources WHERE url = {lit(entry["source_page_url"])});

INSERT INTO politician_photos (politician_id, source_id, source_url, file_path, sha256, bytes, content_type, retrieved_at)
SELECT p.id, s.id, {lit(entry["photo_url"])}, {lit(file_path)}, {lit(digest)}, {len(data)}, {lit(content_type)}, {lit(now)}
FROM politicians p
JOIN sources s ON s.url = {lit(entry["source_page_url"])}
WHERE p.slug = {lit(slug)}
ON CONFLICT (politician_id) DO UPDATE SET
  source_id = EXCLUDED.source_id, source_url = EXCLUDED.source_url, file_path = EXCLUDED.file_path,
  sha256 = EXCLUDED.sha256, bytes = EXCLUDED.bytes, content_type = EXCLUDED.content_type,
  retrieved_at = EXCLUDED.retrieved_at;

UPDATE politicians SET photo_url = {lit("/" + file_path)} WHERE slug = {lit(slug)};
"""


def validate(entry):
    for key in ("slug", "photo_url", "source_name", "source_page_url", "source_type"):
        if not entry.get(key):
            raise ValueError(f"missing {key}")
    if not SLUG.match(entry["slug"]):
        raise ValueError("slug is not in slug form")
    if entry["source_type"] not in ("official", "party", "politician"):
        raise ValueError("source_type must be official, party or politician")
    if not entry["photo_url"].startswith("https://") or not entry["source_page_url"].startswith("https://"):
        raise ValueError("photo and source must be https")


def main(manifest_path):
    entries = json.load(open(manifest_path, encoding="utf-8"))
    os.makedirs(PHOTO_DIR, exist_ok=True)
    log = json.load(open(LOG)) if os.path.exists(LOG) else {}
    saved = failed = 0
    statements = []
    for entry in entries:
        slug = entry.get("slug", "?")
        try:
            validate(entry)
            data = download(entry["photo_url"])
            kind = image_kind(data)
            if kind is None:
                raise ValueError("the file is not a JPEG or PNG image")
            ext, content_type = kind
            digest = hashlib.sha256(data).hexdigest()
            with open(os.path.join(PHOTO_DIR, f"{slug}.{ext}"), "wb") as f:
                f.write(data)
            statements.append(record_sql(entry, data, ext, content_type, digest))
            log[slug] = {"status": "ok", "source_page_url": entry["source_page_url"], "sha256": digest}
            saved += 1
        except (urllib.error.URLError, ValueError, OSError) as error:
            log[slug] = {"status": "failed", "error": str(error), "photo_url": entry.get("photo_url")}
            failed += 1
        time.sleep(0.5)

    if statements:
        script = "BEGIN;\n" + "\n".join(statements) + "\nCOMMIT;\n"
        with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False, encoding="utf-8") as f:
            f.write(script)
            path = f.name
        try:
            subprocess.run(["psql", os.environ["DATABASE_URL"], "-v", "ON_ERROR_STOP=1", "-q", "-f", path], check=True)
        finally:
            os.remove(path)

    with open(LOG, "w", encoding="utf-8") as f:
        json.dump(log, f, indent=1, ensure_ascii=False)
    print(f"saved={saved} failed={failed}")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("usage: python tools/ingest/photos.py <manifest.json>")
    sys.exit(main(sys.argv[1]))
