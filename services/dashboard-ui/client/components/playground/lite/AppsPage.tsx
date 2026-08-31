import { Block } from './Block'
import { LinkBlock } from './LinkBlock'
import { Toolbar } from './Toolbar'
import { apps } from './fixtures'
import { rowHoverClass } from './utils'

const gridClass =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)] gap-5 items-center'

const rowClass = `${gridClass} ${rowHoverClass}`

const headers = ['name & id', 'platform', 'created at', 'updated at']

export const AppsPage = () => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={['Platform']} />

    <div className="flex flex-col gap-3">
      <div className={`${gridClass} pb-1`}>
        {headers.map((header) => (
          <Block
            key={header}
            className="w-[56px] h-[8px] opacity-50"
            title={header}
            text={header}
          />
        ))}
      </div>

      {apps.map((app) => (
        <div key={app.id} className={rowClass}>
          <div className="flex min-w-0 flex-col gap-1.5">
            <LinkBlock
              path={`/apps/${app.id}`}
              label={app.id}
              className="h-[12px] max-w-full"
              style={{ width: app.nameWidth }}
            />
            <Block className="w-[96px] h-[8px] opacity-50" title={app.id} />
          </div>
          <Block
            className="h-[16px] max-w-full rounded-full"
            style={{ width: app.platformWidth }}
          />
          <Block className="w-[88px] h-[10px] max-w-full opacity-60" />
          <Block className="w-[72px] h-[10px] max-w-full opacity-60" />
        </div>
      ))}
    </div>
  </div>
)
