import { memo, useMemo, useState } from 'react'

import { Panel } from '@/components/surfaces/Panel'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'

import type { PlanGroupInstall } from './DeploymentPlanGraph'

interface IDeploymentPlanGroupPanel {
  panelKey: string
  groupName: string
  installs: PlanGroupInstall[]
  orgId: string
  maxParallel: number
  useForPreviews: boolean
  labelEntries: [string, string][]
}

const DeploymentPlanGroupContent = memo(({
  installs,
  orgId,
  maxParallel,
  useForPreviews,
  labelEntries,
}: Omit<IDeploymentPlanGroupPanel, 'panelKey' | 'groupName'>) => {
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
      <div className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <Text variant="subtext" theme="neutral">
            {installs.length} {installs.length === 1 ? 'install' : 'installs'}
          </Text>
          {maxParallel > 1 && (
            <Text variant="subtext" theme="neutral">
              · {maxParallel} in parallel
            </Text>
          )}
          {useForPreviews && (
            <Text variant="subtext" theme="neutral">
              · used for previews
            </Text>
          )}
        </div>
        {labelEntries.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {labelEntries.map(([k, v]) => (
              <LabelBadge key={k} labelKey={k} labelValue={v} size="xs" />
            ))}
          </div>
        )}
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
        <Text theme="neutral">No matching installs</Text>
      ) : filteredInstalls.length === 0 ? (
        <Text theme="neutral">No installs match "{query}"</Text>
      ) : (
        <div className="grid grid-cols-1 @lg:grid-cols-2 @4xl:grid-cols-3 gap-3">
          {filteredInstalls.map((inst) => {
            const installLabels = Object.entries(inst.labels ?? {})
            return (
              <Card key={inst.id} className="!p-4 !gap-3">
                <div className="flex flex-col gap-1 min-w-0">
                  <Text weight="strong" className="truncate" title={inst.name}>
                    {inst.name}
                  </Text>
                  <ID>{inst.id}</ID>
                </div>

                {installLabels.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {installLabels.map(([k, v]) => (
                      <LabelBadge key={k} labelKey={k} labelValue={v} size="xs" />
                    ))}
                  </div>
                )}

                <Link
                  href={`/${orgId}/installs/${inst.id}`}
                  className="flex w-fit items-center gap-1"
                >
                  View install
                  <Icon variant="ArrowRightIcon" size={14} />
                </Link>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
})

DeploymentPlanGroupContent.displayName = 'DeploymentPlanGroupContent'

export const DeploymentPlanGroupPanel = ({
  panelKey,
  groupName,
  installs,
  orgId,
  maxParallel,
  useForPreviews,
  labelEntries,
}: IDeploymentPlanGroupPanel) => (
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
    <DeploymentPlanGroupContent
      installs={installs}
      orgId={orgId}
      maxParallel={maxParallel}
      useForPreviews={useForPreviews}
      labelEntries={labelEntries}
    />
  </Panel>
)
