export default {
  title: 'Branches/WorkflowStepDetail/BuildStep',
}

import type { ReactNode } from 'react'
import { BuildRow, BuildRowDetail, BuildStep } from './BuildStep'
import { AppContext } from '@/providers/app-provider'
import type { TCompositeError } from '@/types'

const RowList = ({ children }: { children: ReactNode }) => (
  <div className="border rounded-[10px] divide-y overflow-hidden">
    {children}
  </div>
)

const mockApp = { id: 'app-1', name: 'demo-app' } as any

const buildError: TCompositeError = {
  type: 'build.source_fetch_failed',
  severity: 'error',
  message: 'Unable to fetch the build source',
  sections: [
    {
      heading: 'How to fix',
      body: 'Check that the repository and ref exist, then retry the build.',
    },
  ],
}

const WithApp = ({ children }: { children: ReactNode }) => (
  <AppContext.Provider
    value={{ app: mockApp, labelColors: {}, refresh: () => {} }}
  >
    {children}
  </AppContext.Provider>
)

export const Rows = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="terraform_module"
      build={{
        component_id: 'c1',
        component_name: 'rds_subnet',
        status: 'success',
        change_reason: 'source_changed',
      }}
    />
    <BuildRow
      rowId="c2"
      type="helm_chart"
      build={{
        component_id: 'c2',
        component_name: 'coder',
        status: 'skipped',
        change_reason: 'no_changes',
      }}
    />
    <BuildRow
      rowId="c3"
      type="docker_build"
      build={{
        component_id: 'c3',
        component_name: 'api',
        status: 'in-progress',
        change_reason: 'config_changed',
      }}
    />
    <BuildRow
      rowId="c4"
      type="kubernetes_manifest"
      build={{
        component_id: 'c4',
        component_name: 'migrations',
        status: 'error',
        change_reason: 'source_changed',
      }}
    />
    <BuildRow
      rowId="sandbox"
      build={{
        component_id: 'sandbox',
        component_type: 'sandbox',
        component_name: 'Sandbox',
        status: 'skipped',
        change_reason: 'no_changes',
      }}
    />
  </RowList>
)

export const Deployed = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="terraform_module"
      build={{
        component_id: 'c1',
        component_name: 'rds_subnet',
        status: 'success',
        cache_status: 'partial cache',
      }}
    />
  </RowList>
)

export const InProgress = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="docker_build"
      build={{
        component_id: 'c1',
        component_name: 'api',
        status: 'in-progress',
        cache_status: 'no cache',
      }}
    />
  </RowList>
)

export const NoType = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      build={{
        component_id: 'c1',
        component_name: 'unknown-component',
        status: 'success',
      }}
    />
  </RowList>
)

export const StartingBuilds = () => (
  <WithApp>
    <BuildStep metadata={{ builds: [] }} status="in-progress" />
  </WithApp>
)

export const WaitingToStartBuilds = () => (
  <WithApp>
    <BuildStep metadata={{ builds: [] }} status="pending" />
  </WithApp>
)

export const ComponentBuildError = () => (
  <BuildRowDetail
    detail={{
      id: 'build-1',
      org_id: 'org-1',
      status: 'error',
      created_at: '2026-08-11T10:00:00Z',
      updated_at: '2026-08-11T10:01:00Z',
      composite_error: buildError,
    }}
    buildHref="/org-1/apps/app-1/components/c1/builds/build-1"
  />
)

export const SandboxBuildError = () => (
  <BuildRowDetail
    detail={{
      id: 'sandbox-build-1',
      status: 'error',
      created_at: '2026-08-11T10:00:00Z',
      updated_at: '2026-08-11T10:01:00Z',
      composite_error: buildError,
    }}
    buildHref="/org-1/apps/app-1/sandbox/builds/sandbox-build-1"
  />
)

export const LoadingBuildDetails = () => <BuildRowDetail isLoading />

export const MissingBuildDetails = () => <BuildRowDetail />
