import { Block } from './Block'
import { Panel } from './Panel'

const paragraph = ['96%', '88%', '92%', '64%']

const Paragraph = () => (
  <div className="flex flex-col gap-2">
    {paragraph.map((width, i) => (
      <Block key={i} className="h-[10px] opacity-60" style={{ width }} />
    ))}
  </div>
)

export const InstallReadme = () => (
  <div className="flex max-w-[760px] flex-col gap-8">
    <div className="flex flex-col gap-4">
      <Block className="h-[22px] w-[280px]" title="heading" />
      <Paragraph />
    </div>

    <div className="flex flex-col gap-4">
      <Block className="h-[16px] w-[200px]" title="section heading" />
      <Paragraph />
    </div>

    <Panel title="Terminal">
      <div className="flex flex-col gap-2">
        {['72%', '54%', '81%'].map((width, i) => (
          <Block key={i} className="h-[10px] opacity-70" style={{ width }} />
        ))}
      </div>
    </Panel>

    <div className="flex flex-col gap-4">
      <Block className="h-[16px] w-[240px]" title="section heading" />
      <div className="flex flex-col gap-3">
        {['86%', '78%', '90%'].map((width, i) => (
          <div key={i} className="flex items-center gap-3">
            <Block className="h-[8px] w-[8px] flex-none rounded-full" />
            <Block className="h-[10px] opacity-60" style={{ width }} />
          </div>
        ))}
      </div>
    </div>
  </div>
)
