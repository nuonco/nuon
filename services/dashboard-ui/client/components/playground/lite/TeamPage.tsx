import { Block } from './Block'
import { Toolbar } from './Toolbar'
import { rowHoverClass } from './utils'

const gridClass =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,1fr)_100px_minmax(0,1fr)_32px] gap-5 items-center'

const headers = ['member', 'role', 'status', 'joined']

const members = [
  { id: 'usr-01', label: 'ana@example.com', role: 116, pending: false },
  { id: 'usr-02', label: 'ben@example.com', role: 92, pending: false },
  { id: 'usr-03', label: 'cass@example.com', role: 104, pending: true },
  { id: 'usr-04', label: 'dev@example.com', role: 88, pending: false },
  { id: 'usr-05', label: 'eli@example.com', role: 120, pending: true },
  { id: 'usr-06', label: 'fran@example.com', role: 96, pending: false },
]

export const TeamPage = () => (
  <div className="flex flex-col gap-4">
    <Toolbar filters={['Role', 'Status']} />

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
        <span />
      </div>

      {members.map((member) => (
        <div key={member.id} className={`${gridClass} ${rowHoverClass}`}>
          <div className="flex min-w-0 items-center gap-3">
            <Block className="h-[24px] w-[24px] flex-none rounded-full" />
            <div className="flex min-w-0 flex-col gap-1.5">
              <Block
                className="h-[12px] max-w-full"
                text={member.label}
                style={{ width: 170 }}
              />
              <Block
                className="h-[8px] w-[96px] max-w-full opacity-50"
                text={member.id}
              />
            </div>
          </div>

          <Block
            className="h-[10px] max-w-full opacity-70"
            style={{ width: member.role }}
          />

          <Block
            className="h-[16px] w-[72px] max-w-full rounded-full"
            title={member.pending ? 'Invited' : 'Active'}
            text={member.pending ? 'Invited' : 'Active'}
          />

          <Block
            className="h-[10px] w-[88px] max-w-full opacity-60"
            text={member.pending ? 'Pending' : undefined}
          />

          <Block
            className="h-[16px] flex-none cursor-pointer opacity-50"
            icon="DotsThreeIcon"
            iconSize={16}
          />
        </div>
      ))}
    </div>
  </div>
)
