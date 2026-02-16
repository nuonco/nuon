# Create AdHoc Action

Creates and executes a one-time action on an install without creating a permanent action workflow definition.

## Use Cases

- Running ad-hoc debugging scripts
- Executing maintenance commands
- Testing bash snippets before creating permanent actions
- Quick data exports or transformations

## Request Body

Provide **either** `inline_contents` (for multi-line bash scripts) **or** `command` (for single-line commands), but not both.

### Fields

- `inline_contents` (string, optional): Multi-line bash script to execute
- `command` (string, optional): Single-line shell command to execute
- `env_vars` (object, optional): Environment variables as key-value pairs
- `timeout` (integer, optional): Execution timeout in seconds (1-3600, default: 300)
- `name` (string, optional): Display name for the action (max 255 chars)

## Response

Returns the created adhoc action run with status information.

## Example

```bash
curl -X POST https://api.nuon.co/v1/installs/{install_id}/actions/adhoc \
  -H "Authorization: Bearer $API_KEY" \
  -H "X-Nuon-Org-ID: $ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "inline_contents": "#!/bin/bash\necho \"Hello from adhoc action\"\nenv | grep NUON",
    "env_vars": {
      "DEBUG": "true",
      "LOG_LEVEL": "info"
    },
    "timeout": 300,
    "name": "Debug Script"
  }'
```

## Notes

- AdHoc actions are marked with `trigger_type: "adhoc"`
- They appear in action run history and can be filtered via trigger_type
- Execution happens on the install's runner using the same infrastructure as permanent actions
- Logs are preserved and can be viewed via the action runs API
- AdHoc runs are kept indefinitely (same retention as regular action runs)
