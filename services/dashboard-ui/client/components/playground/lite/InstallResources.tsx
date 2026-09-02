import { Block } from './Block'
import { Toolbar } from './Toolbar'
import { resources } from './fixtures'
import { rowHoverClass } from './utils'

const gridClass =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,1.4fr)_100px_minmax(0,1fr)] gap-5 items-center'

const rowClass = `${gridClass} ${rowHoverClass}`

const headers = ['name', 'type', 'status', 'updated at']

export const InstallResources = () => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={['Type']} />

    <div className="flex flex-col gap-3">
      <div className={`${gridClass} pb-1`}>
        {headers.map((header) => (
          <Block
            key={header}
            className="h-[8px] w-[56px] max-w-full opacity-50"
            title={header}
            text={header}
          />
        ))}
      </div>

      {resources.map((resource) => (
        <div key={resource.id} className={rowClass}>
          <div className="flex min-w-0 flex-col gap-1.5">
            <Block
              className="h-[12px] max-w-full"
              style={{ width: resource.nameWidth }}
            />
            <Block
              className="h-[8px] w-[110px] max-w-full opacity-50"
              title={resource.id}
            />
          </div>
          <Block
            className="h-[10px] max-w-full opacity-70"
            style={{ width: resource.typeWidth }}
          />
          <Block className="h-[16px] w-[64px] max-w-full rounded-full" />
          <Block className="h-[10px] w-[88px] max-w-full opacity-60" />
        </div>
      ))}
    </div>
  </div>
)
