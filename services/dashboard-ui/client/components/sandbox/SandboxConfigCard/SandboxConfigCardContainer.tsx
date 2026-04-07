import { useSurfaces } from '@/hooks/use-surfaces'
import {
  SandboxEnvironmentVariablesModal,
  SandboxVariablesFilesModal,
} from '@/components/sandbox/SandboxConfigModals'
import { SandboxConfigCard } from './SandboxConfigCard'
import type { ICard } from '@/components/common/Card'
import type { TSandboxConfig } from '@/types'

interface ISandboxConfigCardContainer extends Omit<ICard, 'children'> {
  config: TSandboxConfig
}

export const SandboxConfigCardContainer = ({
  config,
  ...props
}: ISandboxConfigCardContainer) => {
  const { addModal } = useSurfaces()

  const hasEnvVars = config.env_vars && Object.keys(config.env_vars).length > 0
  const hasVariablesFiles = config.variables_files && config.variables_files.length > 0

  const handleViewEnvVars = () => {
    addModal(<SandboxEnvironmentVariablesModal envVars={config.env_vars!} />)
  }

  const handleViewVariablesFiles = () => {
    addModal(<SandboxVariablesFilesModal variablesFiles={config.variables_files!} />)
  }

  return (
    <SandboxConfigCard
      config={config}
      onViewEnvVars={hasEnvVars ? handleViewEnvVars : undefined}
      onViewVariablesFiles={hasVariablesFiles ? handleViewVariablesFiles : undefined}
      {...props}
    />
  )
}
