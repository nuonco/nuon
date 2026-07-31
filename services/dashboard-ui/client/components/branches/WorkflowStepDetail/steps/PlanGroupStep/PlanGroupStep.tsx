import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { useApp } from '@/hooks/use-app'

interface IPlanGroupStep {
  installs: any[]
  groupName?: string
  orgId: string
  labelColors?: Record<string, string>
  hasResponse: boolean
  responseType?: string
  showApproveBar: boolean
  isInProgress: boolean
  actions?: ReactNode
}

export const PlanGroupStep = ({
  installs,
  groupName,
  orgId: _orgId,
  labelColors: _labelColors,
  hasResponse,
  responseType,
  showApproveBar,
  isInProgress: _isInProgress,
  actions,
}: IPlanGroupStep) => {
  const { app } = useApp()

  return (
    <div className="flex flex-col gap-3">
      {hasResponse && (
        <Banner theme="success">
          <Text weight="strong">
            Plan {responseType === 'approve' ? 'approved' : responseType || 'responded'}
          </Text>
        </Banner>
      )}

      <Text variant="label" theme="neutral">
        {groupName || 'Install group'} ({installs.length} {installs.length === 1 ? 'install' : 'installs'})
      </Text>

      {installs.map((inst) => {
        const installName = inst.install_name || inst.install_id
        const newConfigId = inst.new_app_config_id
        const oldConfigId = inst.old_app_config_id

        return (
          <Expand
            key={inst.install_id}
            id={`plan-install-${inst.install_id}`}
            className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg overflow-hidden"
            headerClassName="px-4 py-3"
            heading={
              <div className="flex items-center gap-2 w-full">
                <Text variant="subtext" weight="strong">{installName}</Text>
                {inst.install_labels && Object.entries(inst.install_labels).map(([k, v]) => (
                  <Badge key={k} size="sm" theme="neutral">{k}: {String(v)}</Badge>
                ))}
              </div>
            }
          >
            <div className="p-4 border-t border-cool-grey-100 dark:border-dark-grey-800">
              {newConfigId && app?.id ? (
                <AppConfigDiff
                  appConfigId={newConfigId}
                  oldConfigId={oldConfigId}
                  appId={app.id}
                />
              ) : (
                <Text variant="subtext" theme="neutral">No config changes</Text>
              )}
            </div>
          </Expand>
        )
      })}

      {showApproveBar && (
        <Banner className="@container" theme="warn">
          <div className="flex flex-col gap-2">
            <div className="flex flex-col">
              <Text weight="strong">Install group plan requires review</Text>
              <Text variant="subtext" theme="neutral">
                Review the changes above, then approve to deploy or skip this install group.
              </Text>
            </div>
            {actions && <div className="flex self-end gap-2">{actions}</div>}
          </div>
        </Banner>
      )}
    </div>
  )
}
