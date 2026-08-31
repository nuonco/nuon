import { Block } from './Block'
import { LinkBlock } from './LinkBlock'
import { Toolbar } from './Toolbar'
import { runRows } from './fixtures'
import { rowHoverClass } from './utils'

export interface IRunHistory {
  filters?: string[]
  basePath?: string
}

export const RunHistory = ({ filters = ['Status'], basePath }: IRunHistory) => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={filters} />

    <div className="flex flex-col gap-2">
      {runRows.map((width, i) => {
        const runId = `run-${String(i + 1).padStart(2, '0')}`

        return (
          <div key={runId} className={`flex items-center gap-4 ${rowHoverClass}`}>
            <Block className="h-[16px] w-[16px] flex-none rounded-full" />
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              {basePath ? (
                <LinkBlock
                  path={`${basePath}/${runId}`}
                  label={runId}
                  className="h-[12px]"
                  style={{ width }}
                />
              ) : (
                <Block className="h-[12px]" style={{ width }} />
              )}
              <div className="flex items-center gap-3">
                <Block className="h-[8px] w-[120px] opacity-50" />
                <Block className="h-[14px] w-[88px] rounded-full opacity-70" />
                <Block className="h-[8px] w-[48px] opacity-50" />
              </div>
            </div>
            <Block className="h-[10px] w-[110px] flex-none opacity-50" />
          </div>
        )
      })}
    </div>
  </div>
)
