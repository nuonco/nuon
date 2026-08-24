import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

import { statusAccent } from '../graph/accents'
import type { GroupRunInstall } from './RunDeploymentGraph'

interface IGroupRunDetailPanel extends IPanel {
  groupName: string
  installs: GroupRunInstall[]
  orgId: string
}

export const GroupRunDetailPanel = ({
  groupName,
  installs,
  orgId,
  ...props
}: IGroupRunDetailPanel) => (
  <Panel heading={groupName} {...props}>
    <div className="flex flex-col gap-4 p-4">
      {installs.length === 0 ? (
        <Text theme="neutral">No installs</Text>
      ) : (
        installs.map((inst) => (
          <div key={inst.id} className="flex flex-col gap-1.5">
            <div className="flex items-center gap-2 min-w-0">
              <span className={cn('h-2 w-2 shrink-0 rounded-full', statusAccent(inst.status).dot)} />
              <Link
                href={
                  inst.workflowId
                    ? `/${orgId}/installs/${inst.id}/workflows/${inst.workflowId}`
                    : `/${orgId}/installs/${inst.id}`
                }
                className="min-w-0 flex-1 truncate font-strong"
                title={inst.name}
              >
                {inst.name}
              </Link>
            </div>
            {inst.runbooks.length > 0 && (
              <div className="flex flex-col gap-1 pl-4">
                {inst.runbooks.map((rb) => (
                  <div key={rb.name} className="flex items-center gap-2 min-w-0">
                    <span className={cn('h-1.5 w-1.5 shrink-0 rounded-full', statusAccent(rb.status).dot)} />
                    <Text variant="subtext" theme="neutral" className="truncate">
                      {rb.name}
                    </Text>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))
      )}
    </div>
  </Panel>
)
