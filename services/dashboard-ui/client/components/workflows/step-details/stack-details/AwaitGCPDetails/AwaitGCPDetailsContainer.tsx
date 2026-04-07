import { useInstall } from '@/hooks/use-install'
import { AwaitGCPDetails } from './AwaitGCPDetails'
import type { IStackDetails } from '../types'

export const AwaitGCPDetailsContainer = ({ stack, step }: IStackDetails) => {
  const { install } = useInstall()
  return <AwaitGCPDetails stack={stack} step={step} installId={install?.id} />
}
