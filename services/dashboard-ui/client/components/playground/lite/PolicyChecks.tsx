import { Block } from './Block'
import { Panel } from './Panel'
import { policyChecks } from './fixtures'
import { rowHoverClass } from './utils'

const failed = policyChecks.filter((check) => !check.passed).length

export const PolicyChecks = () => (
  <Panel
    title="Policy checks"
    action={failed > 0 ? `${failed} failed` : 'All passed'}
  >
    <div className="flex flex-col gap-1">
      {policyChecks.map((check) => (
        <div
          key={check.id}
          className={`flex items-center gap-4 ${rowHoverClass}`}
          title={check.label}
        >
          <Block
            className="h-[14px] flex-none"
            icon={check.passed ? 'CheckCircleIcon' : 'WarningCircleIcon'}
            iconSize={14}
          />
          <div className="flex min-w-0 flex-1 flex-col gap-1.5">
            <Block
              className="h-[10px]"
              text={check.label}
              style={{ width: 160 }}
            />
            <Block className="h-[8px] w-[96px] opacity-50" text={check.id} />
          </div>
          <Block
            className="h-[14px] w-[72px] flex-none rounded-full opacity-70"
            text={check.severity}
          />
        </div>
      ))}
    </div>
  </Panel>
)
