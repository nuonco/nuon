import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { IModal } from '@/components/surfaces/Modal'
import type { IPanel } from '@/components/surfaces/Panel'

export const ModalStory = ({
  children,
  label,
}: {
  children: ReactNode
  label?: string
}) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="primary"
      onClick={() => addModal(children as React.ReactElement<IModal>)}
    >
      {label || 'Open modal'}
    </Button>
  )
}

export const PanelStory = ({
  children,
  label,
}: {
  children: ReactNode
  label?: string
}) => {
  const { addPanel } = useSurfaces()

  return (
    <Button
      variant="primary"
      onClick={() => addPanel(children as React.ReactElement<IPanel>)}
    >
      {label || 'Open panel'}
    </Button>
  )
}
