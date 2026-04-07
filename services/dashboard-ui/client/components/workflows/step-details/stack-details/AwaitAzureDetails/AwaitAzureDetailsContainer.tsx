import { useInstall } from '@/hooks/use-install'
import { AwaitAzureDetails } from './AwaitAzureDetails'
import type { IStackDetails } from '../types'

export const AwaitAzureDetailsContainer = ({ stack, step }: IStackDetails) => {
  const { install } = useInstall()

  return (
    <AwaitAzureDetails
      stack={stack}
      step={step}
      installId={install?.id ?? ''}
      azureLocation={install?.azure_account?.location}
    />
  )
}
