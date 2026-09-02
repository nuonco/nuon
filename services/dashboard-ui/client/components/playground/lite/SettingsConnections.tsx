import { Block } from './Block'
import { Panel } from './Panel'
import { rowHoverClass } from './utils'

const githubConnections = [
  { id: 'ghc-01', label: 'acme', meta: 118 },
  { id: 'ghc-02', label: 'acme-labs', meta: 92 },
]

const slackConnections = [{ id: 'slk-01', label: 'acme.slack.com', meta: 134 }]

const ConnectionRows = ({
  rows,
  icon,
}: {
  rows: { id: string; label: string; meta: number }[]
  icon: 'GitHub' | 'SlackLogoIcon'
}) => (
  <div className="flex flex-col gap-2">
    {rows.map((row) => (
      <div key={row.id} className={`flex items-center gap-4 ${rowHoverClass}`}>
        <Block className="h-[20px] flex-none" icon={icon} iconSize={20} />
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <Block className="h-[12px]" text={row.label} style={{ width: 140 }} />
          <Block className="h-[8px] opacity-50" style={{ width: row.meta }} />
        </div>
        <Block className="h-[16px] w-[64px] flex-none rounded-full" />
        <Block className="h-[10px] w-[88px] flex-none opacity-50" />
      </div>
    ))}
  </div>
)

export const SettingsConnections = () => (
  <>
    <Panel title="GitHub" action="Connect GitHub">
      <ConnectionRows rows={githubConnections} icon="GitHub" />
    </Panel>

    <Panel title="Slack" action="Connect Slack">
      <ConnectionRows rows={slackConnections} icon="SlackLogoIcon" />
    </Panel>
  </>
)
