import { memo, useMemo, useState } from 'react'

import { Panel } from '@/components/surfaces/Panel'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

import type { GroupRunInstall } from './RunDeploymentGraph'

interface IGroupRunDetailPanel {
  panelKey: string
  groupName: string
  status?: string
  completedInstalls: number
  totalInstalls: number
  installs: GroupRunInstall[]
  orgId: string
}

const GroupRunDetailContent = memo(({
  status,
  completedInstalls,
  totalInstalls,
  installs,
  orgId,
}: Omit<IGroupRunDetailPanel, 'panelKey' | 'groupName'>) => {
  const [query, setQuery] = useState('')

  const filteredInstalls = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return installs
    return installs.filter(
      (inst) =>
        inst.name.toLowerCase().includes(q) || inst.id.toLowerCase().includes(q)
    )
  }, [installs, query])

  return (
    <div className="@container flex flex-col gap-4">
      <div className="flex items-center gap-2">
        {status && <Status status={status} variant="badge" />}
        <Text variant="subtext" theme="neutral">
          {completedInstalls}/{totalInstalls} installs complete
        </Text>
      </div>

      {installs.length > 0 && (
        <SearchInput
          value={query}
          onChange={setQuery}
          onClear={() => setQuery('')}
          placeholder="Search by install name or id"
          className="w-full md:min-w-0"
          labelClassName="w-full"
        />
      )}

      {installs.length === 0 ? (
        <Text theme="neutral">No installs</Text>
      ) : filteredInstalls.length === 0 ? (
        <Text theme="neutral">No installs match "{query}"</Text>
      ) : (
        <div className="grid grid-cols-1 @lg:grid-cols-2 @4xl:grid-cols-3 gap-3">
          {filteredInstalls.map((inst) => (
            <Card key={inst.id} className="!p-4 !gap-3">
              <div className="flex items-start justify-between gap-3 min-w-0">
                <div className="flex flex-col gap-1 min-w-0">
                  <Text weight="strong" className="truncate" title={inst.name}>
                    {inst.name}
                  </Text>
                  <ID>{inst.id}</ID>
                </div>
                <Status status={inst.status} variant="badge" />
              </div>

              {inst.workflowId && (
                <Link
                  href={`/${orgId}/installs/${inst.id}/workflows/${inst.workflowId}`}
                  className="flex w-fit items-center gap-1"
                >
                  View workflow
                  <Icon variant="ArrowRightIcon" size={14} />
                </Link>
              )}

              {inst.runbooks.length > 0 && (
                <div className="flex flex-col gap-2 border-t pt-3">
                  <Text variant="subtext" theme="neutral" weight="strong">
                    Runbooks
                  </Text>
                  {inst.runbooks.map((rb) => (
                    <div
                      key={rb.name}
                      className="flex items-center justify-between gap-2 min-w-0"
                    >
                      <Text variant="subtext" className="truncate">
                        {rb.name}
                      </Text>
                      <Status status={rb.status} variant="badge" />
                    </div>
                  ))}
                </div>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  )
})

GroupRunDetailContent.displayName = 'GroupRunDetailContent'

export const GroupRunDetailPanel = ({
  panelKey,
  groupName,
  status,
  completedInstalls,
  totalInstalls,
  installs,
  orgId,
}: IGroupRunDetailPanel) => (
  <Panel
    heading={groupName}
    panelKey={panelKey}
    size="half"
    triggerButton={{
      children: 'View details',
      variant: 'secondary',
      size: 'sm',
      className: 'nodrag mt-2 w-full justify-center',
    }}
  >
    <GroupRunDetailContent
      status={status}
      completedInstalls={completedInstalls}
      totalInstalls={totalInstalls}
      installs={installs}
      orgId={orgId}
    />
  </Panel>
)
