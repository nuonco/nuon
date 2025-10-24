'use client'

import { SlidersHorizontalIcon } from '@phosphor-icons/react'
import { Dropdown } from '@/components/old/Dropdown'
import { Text } from '@/components/old/Typography'
import type { TRunner, TRunnerSettings } from '@/types'
import { DeprovisionRunnerModal } from './DeprovisionRunnerModal'
import { ShutdownInstanceModal } from './ShutdownInstance'
import { ShutdownRunnerModal } from './ShutdownRunnerModal'
import { UpdateRunnerModal } from './UpdateRunnerModal'

export const ManageRunnerDropdown = ({
  isInstallRunner = false,
  runner,
  settings,
}: {
  isInstallRunner?: boolean
  runner: TRunner
  settings: TRunnerSettings
}) => {
  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      alignment="right"
      id="runner-dropdown"
      text={
        <>
          <SlidersHorizontalIcon size="16" />
          Manage runner
        </>
      }
      isDownIcon
      wrapperClassName="z-20"
    >
      <div className="min-w-[256px] rounded-md overflow-hidden p-2 flex flex-col gap-1">
        <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
          Controls
        </Text>
        {settings ? (
          <UpdateRunnerModal runnerId={runner?.id} settings={settings} />
        ) : null}
        <ShutdownRunnerModal runnerId={runner?.id} />
        {isInstallRunner ? (
          <ShutdownInstanceModal runnerId={runner?.id} />
        ) : null}

        {isInstallRunner ? (
          <>
            <hr className="my-2" />
            <Text className="px-2 pt-2 pb-1 text-cool-grey-600 dark:text-cool-grey-400">
              Remove
            </Text>
            <DeprovisionRunnerModal />
          </>
        ) : null}
      </div>
    </Dropdown>
  )
}
