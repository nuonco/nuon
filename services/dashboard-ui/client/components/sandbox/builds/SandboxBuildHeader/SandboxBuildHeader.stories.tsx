export default {
  title: 'Sandbox/SandboxBuildHeader',
}

import type { TApp, TAppSandboxBuild } from '@/types'
import { SandboxBuildHeader } from './SandboxBuildHeader'

const app = {
  id: 'app-001',
  name: 'acme-payments',
  org_id: 'org-001',
} as unknown as TApp

const build = {
  id: 'asb01hzk8t3fqp2r9x4m7wcn5vb',
  created_at: '2026-08-24T10:12:00Z',
  updated_at: '2026-08-24T10:15:20Z',
  status_v2: {
    status: 'success',
    status_human_description: 'Sandbox build finished.',
  },
  created_by: { email: 'jane@example.com' },
  runner_job: { id: 'rj01hzk8t3fqp2r9x4m7wcn5vb' },
} as unknown as TAppSandboxBuild

export const Default = () => (
  <SandboxBuildHeader app={app} build={build} orgId="org-001" />
)

export const Failed = () => (
  <SandboxBuildHeader
    app={app}
    build={
      {
        ...build,
        status_v2: {
          status: 'error',
          status_human_description: 'Terraform init failed.',
        },
      } as unknown as TAppSandboxBuild
    }
    orgId="org-001"
  />
)
