import { Block } from './Block'
import { Panel } from './Panel'
import { roles } from './fixtures'
import { rowHoverClass } from './utils'

const gridClass =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,1.4fr)_100px] gap-5 items-center'

const headers = ['name & id', 'trust', 'status']

export const RolesPage = () => (
  <Panel title="IAM roles" action="View in cloud">
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

      {roles.map((role) => (
        <div key={role.id} className={`${gridClass} ${rowHoverClass}`}>
          <div className="flex min-w-0 flex-col gap-1.5">
            <Block
              className="h-[12px] max-w-full"
              style={{ width: role.nameWidth }}
            />
            <Block
              className="h-[8px] w-[96px] max-w-full opacity-50"
              text={role.id}
            />
          </div>
          <Block
            className="h-[10px] max-w-full opacity-70"
            style={{ width: role.metaWidth }}
          />
          <Block className="h-[16px] w-[64px] max-w-full rounded-full" />
        </div>
      ))}
    </div>
  </Panel>
)
