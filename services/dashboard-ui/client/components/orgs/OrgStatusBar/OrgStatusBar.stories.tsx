export default {
  title: 'Orgs/OrgStatusBar',
}

import type { ReactNode } from 'react'
import { OrgContext } from '@/providers/org-provider'
import { OrgStatusBar } from './OrgStatusBar'

const org = { id: 'org-1', name: 'My Org' } as any

const orgWithLongName = {
  id: 'org-1',
  name: 'acme-platform | dev',
  vcs_connections: [{ id: 'vcs-1', github_account_name: 'acme-platform' }],
} as any

const app = { id: 'app-1', name: 'acme-enterprise-aws' } as any

const branch = { id: 'abr-1', name: 'eng-sandbox' } as any

const latestConfig = {
  id: 'cfg-1',
  status: 'active',
  created_at: '2026-09-09T12:00:00Z',
} as any

const install = {
  id: 'inst-1',
  name: 'ws-workspace_01m1fbh4ejef2s64yg82j11se9-byo-eks-testing-ground',
  org_id: 'org-1',
  app_id: 'app-1',
  runner_status: 'active',
  sandbox_status: 'active',
  composite_component_status: 'active',
  composite_health_status: 'healthy',
  drifted_objects: [],
} as any

const stack = {
  versions: [{ id: 'stkv-1', composite_status: { status: 'active' } }],
} as any

const approvals = Array.from({ length: 21 }, (_, i) => ({
  id: `apr-${i}`,
  type: 'plan',
})) as any

const activeWorkflows = Array.from({ length: 50 }, (_, i) => ({
  id: `wkf-${i}`,
  type: 'install_deploy',
  status: { status: 'in-progress' },
})) as any

const longNames = {
  org: {
    ...orgWithLongName,
    name: 'acme-platform-engineering | development',
  } as any,
  app: { id: 'app-1', name: 'acme-enterprise-aws-multi-region-app' } as any,
  branch: { id: 'abr-1', name: 'eng-sandbox-preview-environment' } as any,
}

const FullChain = ({
  width,
  names = 'default',
  byoc = false,
}: {
  width: number
  names?: 'default' | 'long'
  byoc?: boolean
}) => {
  const isLong = names === 'long'
  const barOrg = isLong ? longNames.org : orgWithLongName

  return (
    <OrgContext.Provider value={{ org: barOrg, refresh: () => {} }}>
      <div className="flex h-32 flex-col justify-end bg-cool-grey-100 p-4 dark:bg-dark-grey-950">
        <div className="overflow-hidden rounded border" style={{ width }}>
          <OrgStatusBar
            org={barOrg}
            app={isLong ? longNames.app : app}
            branch={isLong ? longNames.branch : branch}
            latestConfig={latestConfig}
            install={install}
            stack={stack}
            approvals={approvals}
            activeWorkflows={activeWorkflows}
            approvalItems={[]}
            workflowItems={[]}
            byocName={byoc ? 'acme payments' : undefined}
            byocColor={byoc ? '#1A6B4A' : undefined}
            byocTextColor={byoc ? '#FFFFFF' : undefined}
          />
        </div>
      </div>
    </OrgContext.Provider>
  )
}

const Bare = ({ children }: { children: ReactNode }) => (
  <div className="flex h-32 flex-col justify-end">{children}</div>
)

export const Default = () => (
  <Bare>
    <OrgStatusBar
      org={org}
      approvals={[]}
      activeWorkflows={[]}
      approvalItems={[]}
      workflowItems={[]}
    />
  </Bare>
)

export const WithBYOCBadge = () => (
  <Bare>
    <OrgStatusBar
      org={org}
      approvals={[]}
      activeWorkflows={[]}
      approvalItems={[]}
      workflowItems={[]}
      byocName="acme payments"
      byocColor="#1A6B4A"
      byocTextColor="#FFFFFF"
    />
  </Bare>
)

export const WithBYOCBadgeLightColor = () => (
  <Bare>
    <OrgStatusBar
      org={org}
      approvals={[]}
      activeWorkflows={[]}
      approvalItems={[]}
      workflowItems={[]}
      byocName="acme"
      byocColor="#F5D90A"
      byocTextColor="#000000"
    />
  </Bare>
)

export const FullContext = () => <FullChain width={1200} />

export const FullContextHalfWidth = () => <FullChain width={720} />

export const FullContextNarrow = () => <FullChain width={560} />

export const FullContextLongNames = () => (
  <FullChain width={1200} names="long" />
)

export const FullContextLongNamesHalfWidth = () => (
  <FullChain width={720} names="long" />
)

export const FullContextWithBYOCBadge = () => <FullChain width={1200} byoc />

export const FullContextWithBYOCBadgeHalfWidth = () => (
  <FullChain width={720} byoc />
)
