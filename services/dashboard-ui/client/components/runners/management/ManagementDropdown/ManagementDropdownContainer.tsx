import { useRunner } from '@/hooks/use-runner'
import { ManagementDropdown } from './ManagementDropdown'
import type { TRunnerSettings } from '@/types'

export const ManagementDropdownContainer = ({
  isInstallRunner = false,
  settings,
}: {
  isInstallRunner?: boolean
  settings: TRunnerSettings
}) => {
  const { runner } = useRunner()
  if (!runner) return null

  return (
    <ManagementDropdown
      runner={runner}
      isInstallRunner={isInstallRunner}
      settings={settings}
    />
  )
}
