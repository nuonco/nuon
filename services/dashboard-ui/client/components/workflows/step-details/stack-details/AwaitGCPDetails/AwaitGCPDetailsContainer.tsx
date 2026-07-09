import { useInstall } from '@/hooks/use-install'
import { AwaitGCPDetails } from './AwaitGCPDetails'
import type { IStackDetails } from '../types'

interface IAwaitGCPDetailsContainer extends IStackDetails {
  spaceliftEnabled?: boolean
  tfProvider?: boolean
}

export const AwaitGCPDetailsContainer = ({
  stack,
  step,
  spaceliftEnabled,
  tfProvider,
}: IAwaitGCPDetailsContainer) => {
  const { install } = useInstall()
  return (
    <AwaitGCPDetails
      stack={stack}
      step={step}
      installId={install?.id}
      spaceliftEnabled={spaceliftEnabled}
      tfProvider={tfProvider}
    />
  )
}
