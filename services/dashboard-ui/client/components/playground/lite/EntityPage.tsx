import type { ReactNode } from 'react'
import { Block } from './Block'
import { Panel } from './Panel'
import { RunHistory } from './RunHistory'
import { StatTile } from './StatTile'

export interface IEntityPage {
  stats: string[]
  historyTitle: string
  basePath?: string
  children?: ReactNode
}

export const EntityPage = ({
  stats,
  historyTitle,
  basePath,
  children,
}: IEntityPage) => (
  <div className="flex flex-col gap-6">
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {stats.map((stat) => (
        <StatTile key={stat} label={stat} />
      ))}
    </div>

    {children}

    <Panel title={historyTitle} action="View all">
      <RunHistory basePath={basePath} />
    </Panel>
  </div>
)

export const StatePanel = () => (
  <Panel title="Terraform state" action="Download">
    <div className="flex flex-col gap-2">
      {['88%', '64%', '92%', '48%', '76%', '58%', '84%'].map((width, i) => (
        <Block key={i} className="h-[8px] opacity-60" style={{ width }} />
      ))}
    </div>
  </Panel>
)
