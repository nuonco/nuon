import { useInstall } from '@/hooks/use-install'
import { AwaitGCPDetails } from './AwaitGCPDetails'
import type { IStackDetails } from '../types'

export const AwaitGCPDetailsContainer = ({ stack }: IStackDetails) => {
  const { install } = useInstall()
  return <AwaitGCPDetails stack={stack} installId={install?.id} />
}
