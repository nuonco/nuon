import { Block } from './Block'
import { LinkBlock } from './LinkBlock'
import { Toolbar } from './Toolbar'
import { installs } from './fixtures'
import { rowHoverClass } from './utils'

const gridClass =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,1.4fr)_80px_minmax(0,1.1fr)_minmax(0,1.4fr)_minmax(0,1fr)] gap-5 items-center'

const rowClass = `${gridClass} ${rowHoverClass}`

const headers = [
  'name & id',
  'app & branch',
  'status',
  'location',
  'labels',
  'created at',
]

export const InstallsPage = () => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={['App', 'Status']} />

    <div className="flex flex-col gap-3">
      <div className={`${gridClass} pb-1`}>
        {headers.map((header) => (
          <Block
            key={header}
            className="w-[56px] h-[8px] max-w-full opacity-50"
            title={header}
            text={header}
          />
        ))}
      </div>

      {installs.map((install) => (
        <div key={install.id} className={rowClass}>
          <div className="flex min-w-0 flex-col gap-1.5">
            <LinkBlock
              path={`/installs/${install.id}`}
              label={install.id}
              className="h-[12px] max-w-full"
              style={{ width: install.nameWidth }}
            />
            <Block
              className="w-[96px] h-[8px] max-w-full opacity-50"
              title={install.id}
            />
          </div>

          <div className="flex min-w-0 flex-col gap-1.5">
            <Block
              className="h-[10px] max-w-full opacity-70"
              style={{ width: install.appWidth }}
            />
            <Block
              className="h-[8px] max-w-full opacity-50"
              style={{ width: install.branchWidth }}
              title="branch"
            />
          </div>

          <Block
            className="w-[64px] h-[16px] max-w-full rounded-full"
            title="status"
          />

          <div className="flex min-w-0 flex-col gap-1.5 items-start">
            <Block
              className="h-[16px] max-w-full rounded-full"
              style={{ width: install.platformWidth }}
              title="platform"
            />
            <Block
              className="h-[8px] max-w-full opacity-50"
              style={{ width: install.regionWidth }}
              title="region"
            />
          </div>

          <div className="flex min-w-0 flex-wrap gap-1.5">
            {install.labels.map((width, i) => (
              <Block
                key={i}
                className="h-[14px] max-w-full rounded-full opacity-70"
                style={{ width }}
              />
            ))}
          </div>

          <Block
            className="w-[88px] h-[10px] max-w-full opacity-60"
            title="created at"
          />
        </div>
      ))}
    </div>
  </div>
)
