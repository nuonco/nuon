import React, { type FC } from 'react'
import { Code, Heading } from '@/components'
import type { TInstallDeployPlan } from '@/types'

export const Plan: FC<{ plan: TInstallDeployPlan }> = ({ plan }) => {
  return (
    <>
      <div className="flex flex-col gap-2">
        <Heading variant="subheading">Rendered variables</Heading>
        <Code>
          {plan.actual?.waypoint_plan?.variables?.variables?.map((v, i) => {
            let variable = null
            if (v?.Actual?.TerraformVariable) {
              variable = (
                <span className="flex" key={i?.toString()}>
                  <b>{v?.Actual?.TerraformVariable?.name}:</b>{' '}
                  {v?.Actual?.TerraformVariable?.value}
                </span>
              )
            }

            if (v?.Actual?.HelmValue) {
              variable = (
                <span className="flex" key={i?.toString()}>
                  <b>{v?.Actual?.HelmValue?.name}:</b>{' '}
                  {v?.Actual?.HelmValue?.value}
                </span>
              )
            }

            return variable
          })}
        </Code>
      </div>
      <div className="flex flex-col gap-2">
        <Heading variant="subheading">Intermediate variables</Heading>
        <Code variant="preformated">
          {JSON.stringify(
            plan.actual?.waypoint_plan?.variables?.intermediate_data,
            null,
            2
          )}
        </Code>
      </div>

      <div className="flex flex-col gap-2">
        <Heading variant="subheading">Job config</Heading>
        <Code variant="preformated">
          {plan.actual?.waypoint_plan?.waypoint_job?.hcl_config}
        </Code>
      </div>
    </>
  )
}
