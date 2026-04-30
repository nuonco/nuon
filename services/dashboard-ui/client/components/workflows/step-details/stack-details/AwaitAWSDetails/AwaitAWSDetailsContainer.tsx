import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { AwaitAWSDetails } from './AwaitAWSDetails'
import type { IStackDetails } from '../types'

export const AwaitAWSDetailsContainer = ({ stack, step }: IStackDetails) => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <AwaitAWSDetails
      stack={stack}
      step={step}
      orgId={org.id}
      installId={install?.id}
      installAwsRegion={install?.aws_account?.region}
      hasTerraformInstaller={!!org?.features?.['terraform-installer']}
    />
  )
}
