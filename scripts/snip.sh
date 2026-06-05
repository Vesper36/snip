#!/usr/bin/env bash
# snip-cli.sh - Snip command-line paste uploader
# Usage: ./snip-cli.sh [file] [--expire 1h] [--lang go] [--password secret] [--burn]

set -euo pipefail

API_URL="${SNIP_API_URL:-http://localhost:53524}"
API_TOKEN="${SNIP_API_TOKEN:-}"
EXPIRE="never"
LANG=""
PASSWORD=""
BURN=""
FILE_PATH=""
TITLE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -e|--expire) EXPIRE="$2"; shift 2 ;;
        -l|--lang)   LANG="$2"; shift 2 ;;
        -p|--password) PASSWORD="$2"; shift 2 ;;
        -b|--burn) BURN="true"; shift ;;
        -t|--title) TITLE="$2"; shift 2 ;;
        -u|--api-url) API_URL="$2"; shift 2 ;;
        -T|--token)  API_TOKEN="$2"; shift 2 ;;
        -h|--help)
            cat <<EOF
snip-cli - Upload pastes to a Snip instance

Usage: snip-cli [file] [options]

Options:
  -e, --expire <duration>   10m, 1h, 1d, 1w, 1m, or never
  -l, --lang <language>     Language for syntax highlighting
  -p, --password <secret>   Protect with password
  -b, --burn                Burn after reading
  -t, --title <title>       Paste title
  -u, --api-url <url>       Snip server URL (default: \$SNIP_API_URL)
  -T, --token <token>       API token (default: \$SNIP_API_TOKEN)
  -h, --help                Show this help

Reads from stdin if no file is given.
EOF
            exit 0 ;;
        -*) echo "Unknown option: $1" >&2; exit 1 ;;
        *) FILE_PATH="$1"; shift ;;
    esac
done

# Build JSON payload with python for safe escaping
PAYLOAD=$(python3 -c "
import json, sys
data = {
    'expires_in': '${EXPIRE}',
    'burn_after_read': '${BURN}' == 'true',
}
if '${LANG}': data['language'] = '${LANG}'
if '${PASSWORD}': data['password'] = '${PASSWORD}'
if '${TITLE}': data['title'] = '${TITLE}'
content = sys.stdin.read()
if '${FILE_PATH}':
    with open('${FILE_PATH}', 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()
    if not '${TITLE}':
        import os
        data['title'] = os.path.basename('${FILE_PATH}')
data['content'] = content
print(json.dumps(data))
")

CMD="curl -s -X POST \"${API_URL}/api/v1/pastes\" -H \"Content-Type: application/json\""
if [[ -n "$API_TOKEN" ]]; then
    CMD="$CMD -H \"Authorization: ${API_TOKEN}\""
fi
CMD="$CMD -d '$PAYLOAD'"

RESP=$(eval $CMD)
URL=$(echo "$RESP" | python3 -c "import json,sys; print(json.load(sys.stdin).get('url',''))" 2>/dev/null || echo "$RESP")

if [[ "$URL" == http* ]]; then
    echo "$URL"
else
    echo "Error: $RESP" >&2
    exit 1
fi
