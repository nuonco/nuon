import { useEffect, type ReactNode } from 'react'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { IModal } from '@/components/surfaces/Modal'

function ModalOpener({ children }: { children: ReactNode }) {
  const { addModal } = useSurfaces()

  useEffect(() => {
    addModal(children as React.ReactElement<IModal>)
  }, [])

  return null
}

export const ModalStory = ({ children }: { children: ReactNode }) => (
  <SurfacesProvider>
    <ModalOpener>{children}</ModalOpener>
  </SurfacesProvider>
)
