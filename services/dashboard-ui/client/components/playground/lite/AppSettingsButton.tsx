import { useState } from 'react'
import { Block } from './Block'
import { Drawer } from './Drawer'

const branchSettings = [
  'Rename branch',
  'Set as default',
  'Sync config',
  'Rotate app token',
  'Disconnect branch',
]

export const AppSettingsButton = () => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      <Block
        className="h-[32px] w-[32px] cursor-pointer"
        title="Settings"
        icon="GearIcon"
        iconSize={20}
        onClick={() => setIsOpen(true)}
      />

      {isOpen && (
        <Drawer title="Branch settings" onClose={() => setIsOpen(false)}>
          <div className="flex flex-col gap-3">
            {branchSettings.map((action) => (
              <div
                key={action}
                className="flex cursor-pointer items-center justify-between gap-4 rounded-lg bg-cool-grey-100 p-3 dark:bg-dark-grey-800"
                title={action}
              >
                <Block className="h-[10px] w-[140px]" text={action} />
                <Block
                  className="h-[10px] flex-none opacity-60"
                  icon="CaretRightIcon"
                  iconSize={12}
                />
              </div>
            ))}
          </div>
        </Drawer>
      )}
    </>
  )
}
