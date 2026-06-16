#!/usr/bin/env bash
# Serve the plan review document and handle comment saving.
# Usage: ./serve.sh
# Then open http://localhost:8899 in your browser.
#
# Comments you add inline will be saved as individual .md files
# in the comments/ directory, readable by Claude.

set -euo pipefail
cd "$(dirname "$0")"
PORT=8899

echo "Starting plan review server on http://localhost:$PORT"
echo "Comments will be saved to: $(pwd)/comments/"
echo ""

python3 -c "
import http.server
import json
import os
import re
from urllib.parse import urlparse

COMMENTS_DIR = 'comments'

class Handler(http.server.SimpleHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/save-comment':
            length = int(self.headers['Content-Length'])
            body = json.loads(self.rfile.read(length))

            comment_id = body.get('id', 'comment-unknown')
            section = body.get('section', 'unknown')
            text = body.get('text', '')
            timestamp = body.get('timestamp', '')

            # Sanitize filename
            safe_id = re.sub(r'[^a-zA-Z0-9_-]', '', comment_id)
            filepath = os.path.join(COMMENTS_DIR, f'{safe_id}.md')

            content = f'''---
section: {section}
timestamp: {timestamp}
id: {comment_id}
---

{text}
'''
            os.makedirs(COMMENTS_DIR, exist_ok=True)
            with open(filepath, 'w') as f:
                f.write(content)

            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            self.wfile.write(json.dumps({'ok': True, 'file': filepath}).encode())
            print(f'  Saved comment: {filepath} (section: {section})')
            return

        self.send_response(404)
        self.end_headers()

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        self.end_headers()

    def log_message(self, format, *args):
        if '/save-comment' in str(args):
            return  # already printed above
        super().log_message(format, *args)

http.server.HTTPServer(('', $PORT), Handler).serve_forever()
"
