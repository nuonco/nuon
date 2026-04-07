export default {
  title: 'Admin/AdminPanel',
}

import { Button } from '@/components/common/Button'
import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminPanel } from './AdminPanel'

export const Default = () => {
  const { addPanel } = useSurfaces()
  return (
    <Button onClick={() => addPanel(<AdminPanel />)}>Open admin panel</Button>
  )
}

export const HalfSize = () => {
  const { addPanel } = useSurfaces()
  return (
    <Button onClick={() => addPanel(<AdminPanel size="half" />)}>
      Open admin panel (half)
    </Button>
  )
}
