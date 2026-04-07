import { useOrg } from '@/hooks/use-org'
import { ActionRunHeader } from './ActionRunHeader'
import type { IActionRunHeader } from '../types'

export const ActionRunHeaderContainer = (props: IActionRunHeader) => {
  const { org } = useOrg()
  return <ActionRunHeader {...props} orgId={org.id} />
}
