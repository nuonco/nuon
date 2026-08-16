import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { AwaitAWSDetails } from './AwaitAWSDetails'
import type { IStackDetails } from '../types'

export const AwaitAWSDetailsContainer = ({
  stack,
  step,
  loading,
}: IStackDetails) => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <AwaitAWSDetails
      stack={stack}
      step={step}
      loading={loading}
      orgId={org.id}
      installId={install?.id}
      installAwsRegion={install?.aws_account?.region}
      tfProvider={!!org?.features?.['stack-tf-provider']}
    />
  )
}
