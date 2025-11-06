'use client'

import { Plan } from '@/components/approvals/Plan'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TDeploy } from '@/types'
import type { IStepDetails } from '../types'
import { DeployApply } from './DeployApply'

export const DeployStepDetails = ({ step }: IStepDetails) => {
  const { org } = useOrg()
  const { data: deploy, isLoading } = useQuery<TDeploy>({
    path: `/api/orgs/${org.id}/installs/${step?.owner_id}/deploys/${step.step_target_id}`,
  })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        {isLoading || !deploy ? (
          <DeployStepDetailsSkeleton />
        ) : (
          <>
            <Text variant="base" weight="strong">
              {deploy?.component_name} deployment
            </Text>
            <Text variant="subtext">
              <Link
                href={`/${org.id}/installs/${step.owner_id}/components/${deploy?.component_id}`}
              >
                View component <Icon variant="CaretRight" />
              </Link>
            </Text>

            <Text variant="subtext">
              <Link
                href={`/${org.id}/installs/${step.owner_id}/components/${deploy?.component_id}/${deploy?.id}`}
              >
                View deploy <Icon variant="CaretRight" />
              </Link>
            </Text>
          </>
        )}
      </div>
      {step?.execution_type === 'approval' ? (
        <Plan step={step} />
      ) : (
        <DeployApply initDeploy={deploy} step={step} />
      )}
    </div>
  )
}

export const DeployStepDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="180px" />
      <Skeleton height="17px" width="115px" />
      <Skeleton height="17px" width="115px" />
    </>
  )
}
