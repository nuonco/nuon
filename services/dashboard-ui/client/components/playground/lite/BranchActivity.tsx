import { Block } from './Block'
import { Toolbar } from './Toolbar'
import { activityGroups } from './fixtures'
import { labelWidth, rowHoverClass } from './utils'

export const BranchActivity = () => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={['Type', 'Status']} />

    {activityGroups.map((group) => (
      <div key={group.label} className="flex flex-col gap-2">
        <Block
          className="h-[8px] opacity-50"
          style={{ width: labelWidth(group.label) }}
          title={group.label}
          text={group.label}
        />

        {group.rows.map((row, i) => (
          <div key={i} className={`flex items-center gap-4 ${rowHoverClass}`}>
            <Block className="h-[16px] w-[16px] flex-none rounded-full" />
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Block className="h-[12px]" style={{ width: row.title }} />
              <div className="flex items-center gap-3">
                <Block className="h-[8px] w-[120px] opacity-50" />
                <Block className="h-[14px] w-[96px] rounded-full opacity-70" />
                <Block className="h-[8px] w-[56px] opacity-50" />
              </div>
            </div>
            <Block className="h-[10px] w-[110px] flex-none opacity-50" />
          </div>
        ))}
      </div>
    ))}
  </div>
)
