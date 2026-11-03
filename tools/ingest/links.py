"""Record cited profile links and websites for politicians, each with its publisher and a check result.

Input: a JSON list (see links.json). Each entry:
  slug            the politician's slug in the database
  kind            "profile" (a page about them) or "website" (their own site)
  title           what the link is, e.g. "National Assembly profile"
  url             the https address of the page
  source_name     who publishes the page
  publisher_type  "official" (government, parliament, IEBC) or "third_party" (a monitor, a news site)
  found_via       how the link was found (search query or listing), recorded for the audit trail

Each URL is requested once. A 200 response is recorded as verified. Anything else (a redirect loop, an error, no
connection) is still recorded, but the check note says it was not verified and what happened, so the app never
presents an unchecked link as checked.

Usage (from the repo root):  python tools/ingest/links.py tools/ingest/links.json
Requires DATABASE_URL in the environment, and psql on the PATH.
"""
import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request
from datetime import date

USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124.0 Safari/537.36 JaladaIngestion/1.0"
)
SLUG = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")


class NoRedirects(urllib.request.HTTPRedirectHandler):
    """Report redirects instead of following them, so a site that redirects to itself is seen as it is."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def lit(value):
    return "NULL" if value is None else "'" + str(value).replace("'", "''") + "'"


def check(url):
    """Returns (verified, note)."""
    opener = urllib.request.build_opener(NoRedirects)
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT}, method="GET")
    try:
        with opener.open(req, timeout=30) as r:
            return r.status == 200, f"verified: HTTP {r.status} on {date.today().isoformat()}"
    except urllib.error.HTTPError as e:
        return False, f"not verified: HTTP {e.code} on {date.today().isoformat()}"
    except (urllib.error.URLError, OSError) as e:
        return False, f"not verified: {type(e).__name__} on {date.today().isoformat()}"


def validate(entry):
    for key in ("slug", "kind", "title", "url", "source_name", "publisher_type", "found_via"):
        if not entry.get(key):
            raise ValueError(f"missing {key}")
    if not SLUG.match(entry["slug"]):
        raise ValueError("slug is not in slug form")
    if entry["kind"] not in ("profile", "website"):
        raise ValueError("kind must be profile or website")
    if entry["publisher_type"] not in ("official", "third_party"):
        raise ValueError("publisher_type must be official or third_party")
    if not entry["url"].startswith("https://"):
        raise ValueError("url must be https")


def statement(entry, verified, note):
    today = date.today().isoformat()
    note_full = f"{note}. Found via: {entry['found_via']}"
    return f"""
INSERT INTO sources (name, url, type, reliability, last_accessed_at)
SELECT {lit(entry["source_name"])}, {lit(entry["url"])}, 'other',
       {lit("official" if entry["publisher_type"] == "official" else "trusted")}, now()
WHERE NOT EXISTS (SELECT 1 FROM sources WHERE url = {lit(entry["url"])});

INSERT INTO politician_links (politician_id, kind, title, url, source_id, publisher_type, checked_on, check_note)
SELECT p.id, {lit(entry["kind"])}, {lit(entry["title"])}, {lit(entry["url"])}, s.id,
       {lit(entry["publisher_type"])}, DATE {lit(today)}, {lit(note_full)}
FROM politicians p
JOIN sources s ON s.url = {lit(entry["url"])}
WHERE p.slug = {lit(entry["slug"])}
ON CONFLICT (politician_id, url) DO UPDATE SET
  kind = EXCLUDED.kind, title = EXCLUDED.title, source_id = EXCLUDED.source_id,
  publisher_type = EXCLUDED.publisher_type, checked_on = EXCLUDED.checked_on, check_note = EXCLUDED.check_note;
"""


def main(path):
    entries = json.load(open(path, encoding="utf-8"))
    statements, verified_count, unverified_count = [], 0, 0
    for entry in entries:
        try:
            validate(entry)
        except ValueError as e:
            print(f"rejected {entry.get('slug', '?')}: {e}")
            return 1
        ok, note = check(entry["url"])
        verified_count += ok
        unverified_count += not ok
        print(f"{entry['slug']}: {note}")
        statements.append(statement(entry, ok, note))

    if statements:
        script = "BEGIN;\n" + "\n".join(statements) + "\nCOMMIT;\n"
        with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False, encoding="utf-8") as f:
            f.write(script)
            temp = f.name
        try:
            subprocess.run(["psql", os.environ["DATABASE_URL"], "-v", "ON_ERROR_STOP=1", "-q", "-f", temp], check=True)
        finally:
            os.remove(temp)
    print(f"recorded={len(statements)} verified={verified_count} not_verified={unverified_count}")
    return 0


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("usage: python tools/ingest/links.py <links.json>")
    sys.exit(main(sys.argv[1]))
