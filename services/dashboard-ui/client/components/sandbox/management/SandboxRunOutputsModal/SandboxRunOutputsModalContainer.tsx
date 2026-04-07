import { useSurfaces } from '@/hooks/use-surfaces'
import { SandboxRunOutputsModal, SandboxRunOutputsButton } from './SandboxRunOutputsModal'
import type { IButtonAsButton } from '@/components/common/Button'
import type { TSandboxRun } from '@/types'

interface ISandboxRunOutputsButtonContainer extends IButtonAsButton {
  sandboxRun: TSandboxRun
  headingText?: string
}

export const SandboxRunOutputsButtonContainer = ({
  sandboxRun,
  headingText,
  ...props
}: ISandboxRunOutputsButtonContainer) => {
  const { addModal } = useSurfaces()

  const handleOpen = () => {
    addModal(<SandboxRunOutputsModal sandboxRun={sandboxRun} headingText={headingText} />)
  }

  return (
    <SandboxRunOutputsButton
      sandboxRun={sandboxRun}
      headingText={headingText}
      onOpen={handleOpen}
      {...props}
    />
  )
}
