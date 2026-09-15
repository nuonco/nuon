import { Outlet } from 'react-router'
import { useSimpleIA } from '@/hooks/use-simple-ia'
import { useOrg } from '@/hooks/use-org'
import { NotFound } from '@/views/NotFound'

export const SimpleIAGate = () => {
  const { org } = useOrg()
  const hasSimpleIA = useSimpleIA()

  if (!org) return null
  return hasSimpleIA ? <Outlet /> : <NotFound />
}
