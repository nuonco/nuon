import { useOrg } from '@/hooks/use-org'
import { ActionRunMetadata } from './ActionRunMetadata'
import type { IActionRunMetadata } from '../types'

export const ActionRunMetadataContainer = (props: IActionRunMetadata) => {
  const { org } = useOrg()
  return <ActionRunMetadata {...props} orgId={org.id} />
}
