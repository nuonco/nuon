import { Block } from './Block'
import { Panel } from './Panel'
import { labelWidth } from './utils'

const defaultMeta = [
  { label: 'Author', width: 140 },
  { label: 'Committed', width: 110 },
  { label: 'Connected by', width: 150 },
]

export interface ISourceCard {
  meta?: { label: string; width: number }[]
}

export const SourceCard = ({ meta = defaultMeta }: ISourceCard) => (
  <Panel title="Branch" action="View on GitHub">
    <div className="flex flex-wrap items-center gap-3">
      <Block className="h-[20px] flex-none" icon="GitHub" iconSize={20} />
      <Block className="h-[14px]" text="acme/platform" style={{ width: 150 }} />
      <Block
        className="h-[16px] flex-none rounded-full"
        icon="GitBranchIcon"
        iconSize={12}
        text="main"
      />
      <Block className="h-[16px] w-[72px] flex-none rounded-full" />
    </div>

    <div className="flex items-start gap-3">
      <Block
        className="h-[16px] flex-none opacity-60"
        icon="GitCommitIcon"
        iconSize={16}
        text="a1b2c3d"
      />
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <Block className="h-[12px] w-[70%]" />
        <Block className="h-[8px] w-[45%] opacity-50" />
      </div>
    </div>

    <div className="flex flex-col gap-3">
      {meta.map((item) => (
        <div
          key={item.label}
          className="flex items-center justify-between gap-4"
        >
          <Block
            className="h-[8px] opacity-50"
            style={{ width: labelWidth(item.label) }}
            title={item.label}
            text={item.label}
          />
          <Block className="h-[10px]" style={{ width: item.width }} />
        </div>
      ))}
    </div>
  </Panel>
)
