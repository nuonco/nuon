import React, { Suspense, type FC } from 'react'
import { Card, Heading, Plan, Text } from '@/components'
import { getDeployPlan, type IGetDeployPlan } from '@/lib'
import type { TInstallDeployPlan } from '@/types'

export const DeployPlan: FC<IGetDeployPlan> = async (props) => {
  let plan: TInstallDeployPlan
  try {
    plan = await getDeployPlan(props)
  } catch (error) {
    console.log('error?', error)
    return <Text variant="label">No deploy plan to show</Text>
  }
  return <Plan plan={plan} data-testid="deploy-plan" />
}

export const DeployPlanCard: FC<IGetDeployPlan> = (props) => (
  <Card className="flex-1">
    <Heading>Deploy plan</Heading>
    <Suspense fallback="Loading deploy plan...">
      <DeployPlan {...props} />
    </Suspense>
  </Card>
)
