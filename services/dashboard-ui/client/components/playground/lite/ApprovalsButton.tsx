import { useState } from 'react'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { LinkBlock } from './LinkBlock'
import { approvals } from './fixtures'
import { rowHoverClass } from './utils'

export const ApprovalsButton = () => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      <Block
        className="h-[16px] cursor-pointer"
        title={`${approvals.length} workflows awaiting approval`}
        icon="BellIcon"
        iconSize={14}
        text={`${approvals.length} awaiting approval`}
        onClick={() => setIsOpen(true)}
      />

      {isOpen && (
        <Drawer title="Awaiting approval" onClose={() => setIsOpen(false)}>
          <div className="flex flex-col gap-2">
            {approvals.map((approval) => (
              <div
                key={approval.id}
                className={`flex items-center gap-4 ${rowHoverClass}`}
                title={approval.label}
              >
                <Block className="h-[14px] w-[14px] flex-none rounded-full" />

                <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                  <LinkBlock
                    path={`/installs/${approval.installId}/activity`}
                    label={approval.label}
                    className="h-[12px]"
                    style={{ width: 170 }}
                  />
                  <div className="flex items-center gap-3">
                    <Block
                      className="h-[8px] opacity-50"
                      text={approval.installId}
                      style={{ width: 96 }}
                    />
                    <Block
                      className="h-[8px] opacity-50"
                      style={{ width: approval.waitingWidth }}
                    />
                  </div>
                </div>

                <Block
                  className="h-[24px] w-[72px] flex-none cursor-pointer"
                  icon="CheckIcon"
                  iconSize={12}
                  text="Approve"
                />
              </div>
            ))}
          </div>
        </Drawer>
      )}
    </>
  )
}
