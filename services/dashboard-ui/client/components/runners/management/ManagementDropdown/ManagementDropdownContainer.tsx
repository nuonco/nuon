import { useRunner } from '@/hooks/use-runner'
import { ManagementDropdown } from './ManagementDropdown'
import type { TRunnerSettings } from '@/types'

export const ManagementDropdownContainer = ({
  isInstallRunner = false,
  settings,
  hasMngProcess,
  hasInstanceProcess,
}: {
  isInstallRunner?: boolean
  settings: TRunnerSettings
  hasMngProcess?: boolean
  hasInstanceProcess?: boolean
}) => {
  const { runner } = useRunner()
  if (!runner) return null

  return (
    <ManagementDropdown
      runner={runner}
      isInstallRunner={isInstallRunner}
      settings={settings}
      hasMngProcess={hasMngProcess}
      hasInstanceProcess={hasInstanceProcess}
    />
  )
}
