export default {
  title: 'Common/Editor',
}

import { useState } from 'react'
import { Editor } from './Editor'
import { Text } from './Text'

export const Basic = () => {
  const [code, setCode] = useState(`#!/bin/bash

echo "Starting deployment..."
npm install
npm run build
npm run deploy

echo "Deployment complete!"`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Basic editor</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Editor component provides a code editing experience with syntax
          highlighting. It captures input as a string for use in forms.
        </p>
      </div>

      <div className="space-y-4">
        <Editor
          value={code}
          onChange={setCode}
          language="bash"
          placeholder="Enter your code here..."
        />

        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <Text variant="label" weight="strong">
            Current value:
          </Text>
          <Text variant="base" className="font-mono text-xs mt-2">
            {code.length} characters
          </Text>
        </div>
      </div>
    </div>
  )
}

export const Languages = () => {
  const codeExamples = {
    bash: `#!/bin/bash

echo "Starting deployment..."
npm install
npm run build
npm run deploy

echo "Deployment complete!"`,
    json: `{
  "name": "nuon",
  "version": "1.0.0",
  "description": "BYOC platform",
  "main": "index.js"
}`,
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Language support</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Supports bash and JSON syntax highlighting.
        </p>
      </div>

      <div className="space-y-6">
        {Object.entries(codeExamples).map(([lang, code]) => (
          <div key={lang} className="space-y-2">
            <Text variant="label" weight="strong" className="capitalize">
              {lang}
            </Text>
            <Editor
              value={code}
              language={lang as 'bash' | 'json'}
              minHeight={150}
              readOnly
            />
          </div>
        ))}
      </div>
    </div>
  )
}

export const Sizes = () => {
  const sampleCode = `echo "Hello, world!"`

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Height controls</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Use <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">minHeight</code> and{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">maxHeight</code> props
          to control the editor dimensions.
        </p>
      </div>

      <div className="space-y-6">
        <div className="space-y-2">
          <Text variant="label" weight="strong">Small (100px)</Text>
          <Editor value={sampleCode} language="bash" minHeight={100} maxHeight={100} />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">Medium (200px, default)</Text>
          <Editor value={sampleCode} language="bash" minHeight={200} maxHeight={200} />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">Large (400px)</Text>
          <Editor value={sampleCode} language="bash" minHeight={400} maxHeight={400} />
        </div>
      </div>
    </div>
  )
}

export const States = () => {
  const [editableCode, setEditableCode] = useState(`#!/bin/bash
echo "You can edit this"`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Editor states</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The editor supports editable, read-only, disabled, and empty with placeholder states.
        </p>
      </div>

      <div className="space-y-6">
        <div className="space-y-2">
          <Text variant="label" weight="strong">Editable (default)</Text>
          <Editor value={editableCode} onChange={setEditableCode} language="bash" minHeight={120} />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">Read-only</Text>
          <Editor value={`#!/bin/bash\necho "You cannot edit this"`} language="bash" readOnly minHeight={120} />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">Disabled</Text>
          <Editor value={`#!/bin/bash\necho "Disabled state"`} language="bash" disabled minHeight={120} />
        </div>

        <div className="space-y-2">
          <Text variant="label" weight="strong">Empty with placeholder</Text>
          <Editor value="" placeholder="Enter your code here..." language="bash" minHeight={120} />
        </div>
      </div>
    </div>
  )
}

export const LongScript = () => {
  const [code, setCode] = useState(`#!/bin/bash
set -euo pipefail

LOG_DIR="/var/log/deploy"
BACKUP_DIR="/var/backups/app"
DEPLOY_DIR="/opt/app/current"
HEALTH_CHECK_URL="http://localhost:8080/health"
MAX_RETRIES=30
RETRY_INTERVAL=2

mkdir -p "$LOG_DIR" "$BACKUP_DIR"

echo "Starting deployment at $(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [ -d "$DEPLOY_DIR" ]; then
  BACKUP_NAME="backup-$(date +%s).tar.gz"
  echo "Backing up current deployment to $BACKUP_DIR/$BACKUP_NAME"
  tar -czf "$BACKUP_DIR/$BACKUP_NAME" -C "$DEPLOY_DIR" .
  find "$BACKUP_DIR" -name "backup-*.tar.gz" -mtime +7 -delete
  echo "Old backups cleaned up"
fi

echo "Checking system resources..."
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
  echo "ERROR: Disk usage is at \${DISK_USAGE}%. Aborting deployment."
  exit 1
fi

AVAILABLE_MEM=$(free -m | awk 'NR==2 {print $7}')
if [ "$AVAILABLE_MEM" -lt 512 ]; then
  echo "WARNING: Only \${AVAILABLE_MEM}MB memory available"
fi

echo "Pulling latest configuration from parameter store..."
DB_HOST=$(aws ssm get-parameter --name "/app/prod/db-host" --with-decryption --query "Parameter.Value" --output text 2>/dev/null || echo "localhost")
DB_PORT=$(aws ssm get-parameter --name "/app/prod/db-port" --with-decryption --query "Parameter.Value" --output text 2>/dev/null || echo "5432")
DB_NAME=$(aws ssm get-parameter --name "/app/prod/db-name" --with-decryption --query "Parameter.Value" --output text 2>/dev/null || echo "appdb")
REDIS_URL=$(aws ssm get-parameter --name "/app/prod/redis-url" --with-decryption --query "Parameter.Value" --output text 2>/dev/null || echo "redis://localhost:6379")

echo "Running database migrations..."
for migration in /opt/app/migrations/*.sql; do
  if [ -f "$migration" ]; then
    MIGRATION_NAME=$(basename "$migration")
    echo "  Applying migration: $MIGRATION_NAME"
    psql -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -f "$migration" >> "$LOG_DIR/migrations.log" 2>&1
    echo "  Migration $MIGRATION_NAME applied successfully"
  fi
done

echo "Stopping existing services..."
for service in app-worker app-scheduler app-web; do
  if systemctl is-active --quiet "$service" 2>/dev/null; then
    echo "  Stopping $service"
    systemctl stop "$service"
    sleep 1
  fi
done

echo "Installing dependencies..."
cd "$DEPLOY_DIR"
if [ -f "package.json" ]; then
  npm ci --production >> "$LOG_DIR/npm-install.log" 2>&1
fi
if [ -f "requirements.txt" ]; then
  pip install -r requirements.txt >> "$LOG_DIR/pip-install.log" 2>&1
fi

echo "Starting services..."
for service in app-web app-worker app-scheduler; do
  echo "  Starting $service"
  systemctl start "$service"
  sleep 2
  if ! systemctl is-active --quiet "$service"; then
    echo "ERROR: $service failed to start"
    journalctl -u "$service" --no-pager -n 20
    exit 1
  fi
done

echo "Running health checks..."
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTH_CHECK_URL" 2>/dev/null || echo "000")
  if [ "$HTTP_CODE" = "200" ]; then
    echo "Health check passed after $((RETRY_COUNT + 1)) attempts"
    break
  fi
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "  Health check attempt $RETRY_COUNT/$MAX_RETRIES returned HTTP $HTTP_CODE"
  sleep $RETRY_INTERVAL
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo "ERROR: Health check failed after $MAX_RETRIES attempts. Rolling back..."
  LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/backup-*.tar.gz 2>/dev/null | head -1)
  if [ -n "$LATEST_BACKUP" ]; then
    rm -rf "$DEPLOY_DIR"/*
    tar -xzf "$LATEST_BACKUP" -C "$DEPLOY_DIR"
    for service in app-web app-worker app-scheduler; do
      systemctl restart "$service"
    done
    echo "Rollback complete"
  fi
  exit 1
fi

echo "Deployment completed successfully at $(date -u +%Y-%m-%dT%H:%M:%SZ)"`)

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Long script</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          A long bash script to verify the editor handles large content with proper scrolling and editing.
        </p>
      </div>

      <Editor
        value={code}
        onChange={setCode}
        language="bash"
        minHeight={300}
        maxHeight={500}
      />
    </div>
  )
}
