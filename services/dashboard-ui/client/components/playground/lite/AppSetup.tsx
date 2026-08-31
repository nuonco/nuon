import { Block } from './Block'
import { Panel } from './Panel'

const steps = ['Connect GitHub', 'Choose a branch', 'Review config']

export const AppSetup = () => (
  <div className="flex max-w-[640px] flex-col gap-6">
    <Panel title="Set up this app" action="Docs">
      <div className="flex flex-col gap-4">
        {steps.map((step, i) => (
          <div key={step} className="flex items-center gap-4">
            <Block className="h-[24px] w-[24px] flex-none rounded-full" />
            <div className="flex flex-1 flex-col gap-1.5">
              <Block className="h-[12px]" text={step} style={{ width: 160 }} />
              <Block className="h-[8px] w-[60%] opacity-50" />
            </div>
            {i === 0 && (
              <Block
                className="h-[32px] w-[120px] flex-none cursor-pointer"
                icon="GitHub"
                text="Connect"
              />
            )}
          </div>
        ))}
      </div>
    </Panel>
  </div>
)
