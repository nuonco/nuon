export default {
  title: 'Branches/WorkflowStepDetail/BuildStep',
}

import type { ReactNode } from 'react'
import { BuildRow, BuildStep } from './BuildStep'
import { AppContext } from '@/providers/app-provider'

const RowList = ({ children }: { children: ReactNode }) => (
  <div className="border rounded-[10px] divide-y overflow-hidden">
    {children}
  </div>
)

const mockApp = { id: 'app-1', name: 'demo-app' } as any

const WithApp = ({ children }: { children: ReactNode }) => (
  <AppContext.Provider value={{ app: mockApp, labelColors: {}, refresh: () => {} }}>
    {children}
  </AppContext.Provider>
)

export const Rows = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="terraform_module"
      build={{ component_id: 'c1', component_name: 'rds_subnet', status: 'success', cache_status: 'partial cache' }}
    />
    <BuildRow
      rowId="c2"
      type="helm_chart"
      build={{ component_id: 'c2', component_name: 'coder', status: 'success', cache_status: 'cache hit' }}
    />
    <BuildRow
      rowId="c3"
      type="docker_build"
      build={{ component_id: 'c3', component_name: 'api', status: 'in-progress', cache_status: 'no cache' }}
    />
    <BuildRow
      rowId="c4"
      type="kubernetes_manifest"
      build={{ component_id: 'c4', component_name: 'migrations', status: 'error', cache_status: 'no cache' }}
    />
    <BuildRow
      rowId="sandbox"
      build={{ component_id: 'sandbox', component_type: 'sandbox', component_name: 'Sandbox', status: 'success', cache_status: 'no cache' }}
    />
  </RowList>
)

export const Deployed = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="terraform_module"
      build={{ component_id: 'c1', component_name: 'rds_subnet', status: 'success', cache_status: 'partial cache' }}
    />
  </RowList>
)

export const InProgress = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      type="docker_build"
      build={{ component_id: 'c1', component_name: 'api', status: 'in-progress', cache_status: 'no cache' }}
    />
  </RowList>
)

export const NoType = () => (
  <RowList>
    <BuildRow
      rowId="c1"
      build={{ component_id: 'c1', component_name: 'unknown-component', status: 'success' }}
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
