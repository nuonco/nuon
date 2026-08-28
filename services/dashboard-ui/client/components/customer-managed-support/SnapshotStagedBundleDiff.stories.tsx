import type { Story } from '@ladle/react'
import { SnapshotStagedBundleDiff } from './SnapshotStagedBundleDiff'

export const StagedOnly: Story = () => (
  <div className="max-w-5xl p-8">
    <SnapshotStagedBundleDiff
      candidate={{
        schema_version: 1,
        previous_digest: 'sha256:active-bundle-digest',
        staged_at: '2026-08-25T15:30:00Z',
        archive_name: 'aws-lambda-v2.tar.zst',
        archive_size: 24_117_248,
        bundle: {
          schema_version: 1,
          deployment_id: 'demo-install',
          bundle_digest: 'sha256:staged-bundle-digest',
          activated_at: '0001-01-01T00:00:00Z',
        },
        changes: [
          {
            kind: 'component',
            name: 'lambda-api',
            change: 'changed',
            previous_digest: 'sha256:lambda-v1',
            candidate_digest: 'sha256:lambda-v2',
            previous_config_digest: 'sha256:config-v1',
            candidate_config_digest: 'sha256:config-v2',
            previous_component_definition: {
              type: 'terraform_module',
              terraform_module: {
                source: './components/lambda',
                version: '1.0.0',
              },
            },
            candidate_component_definition: {
              type: 'terraform_module',
              terraform_module: {
                source: './components/lambda',
                version: '1.1.0',
              },
            },
          },
          {
            kind: 'stack-asset',
            name: 'root',
            detail: 'compiled:root',
            change: 'changed',
            previous_digest: 'sha256:stack-v1',
            candidate_digest: 'sha256:stack-v2',
          },
          {
            kind: 'runbook',
            name: 'verify-lambda',
            change: 'added',
            candidate_digest: 'sha256:runbook-v1',
            candidate_runbook_definition: {
              steps: [
                { name: 'invoke-function', command: 'aws lambda invoke' },
              ],
            },
          },
        ],
      }}
    />
  </div>
)

StagedOnly.storyName = 'Staged bundle · not deployed'
