import React, { Suspense, type FC } from 'react'
import { Card, Heading, Plan, Text } from '@/components'
import { getBuildPlan, type IGetBuildPlan } from '@/lib'
import type { TComponentBuildPlan } from '@/types'

export const BuildPlan: FC<IGetBuildPlan> = async (props) => {
  let plan: TComponentBuildPlan
  try {
    plan = await getBuildPlan(props)
  } catch (error) {
    return <Text variant="label">No build plan to show</Text>
  }

  return <Plan plan={plan} />
}

export const BuildPlanCard: FC<IGetBuildPlan> = (props) => {
  return (
    <Card className="flex-1">
      <Heading>Build plan</Heading>
      <Suspense fallback="Loading build plan">
        <BuildPlan {...props} />
      </Suspense>
    </Card>
  )
}
