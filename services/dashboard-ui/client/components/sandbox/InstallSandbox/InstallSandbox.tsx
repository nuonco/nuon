import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TInstallSandbox, TSandboxRun } from '@/types'
import { humanize } from '@/utils/string-utils'

type TSandboxRunSummary = Pick<TSandboxRun, 'created_at' | 'run_type'>

export interface IInstallSandbox {
  actions?: ReactNode
  config: ReactNode
  driftBanner?: ReactNode
  latestRun?: TSandboxRunSummary
  loading?: boolean
  sandbox?: TInstallSandbox
}

export const InstallSandbox = ({
  actions,
  config,
  driftBanner,
  latestRun,
  loading = false,
  sandbox,
}: IInstallSandbox) => (
  <div className="flex flex-col gap-6">
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Sandbox"
        description="Infrastructure provisioned for this install."
        status={
          sandbox?.status_v2?.status || sandbox?.status ? (
            <Status
              status={sandbox.status_v2?.status ?? sandbox.status}
              variant="badge"
            />
          ) : undefined
        }
        actions={actions}
      />

      {loading ? (
        <Card className="!p-4">
          <div className="flex flex-wrap gap-x-8 gap-y-3">
            <LabeledValue label="Sandbox ID" loading />
            <LabeledValue label="Run type" loading />
            <LabeledValue label="Last run" loading />
          </div>
        </Card>
      ) : sandbox ? (
        <Card className="!p-4">
          <div className="flex flex-wrap gap-x-8 gap-y-3">
            <LabeledValue label="Sandbox ID">
              <ID>{sandbox.id}</ID>
            </LabeledValue>
            {latestRun?.run_type ? (
              <LabeledValue label="Run type">
                <Badge size="sm" theme="neutral">
                  {humanize(latestRun.run_type)}
                </Badge>
              </LabeledValue>
            ) : null}
            {latestRun?.created_at ? (
              <LabeledValue label="Last run">
                <Time
                  time={latestRun.created_at}
                  format="relative"
                  variant="subtext"
                />
              </LabeledValue>
            ) : null}
            {sandbox.terraform_workspace?.id ? (
              <LabeledValue label="Workspace ID">
                <ID>{sandbox.terraform_workspace.id}</ID>
              </LabeledValue>
            ) : null}
          </div>
        </Card>
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No sandbox provisioned"
          emptyMessage="Provision the sandbox to create infrastructure for this install."
        />
      )}
    </div>

    {driftBanner}
    {config}
  </div>
)
