import React, { FC } from 'react'
import {
  getWorkspaceStates,
  getWorkspaceState,
  getInstallSandboxRuns,
  getWorkspaceStateResources,
} from '@/lib'
import { Section } from '@/components/Card'
import { Time } from '@/components/Time'
import { WorkspaceManagementDropdown } from '@/components/InstallSandbox/WorkspaceManagementDropdown'
import { DataTable } from '@/components/InstallSandbox/Table'
import { Tabs, Tab } from '@/components/InstallSandbox/Tabs'
import { getToken } from '@/components/admin-actions'

export interface ITerraformWorkspace {
  orgId: string
  installId: string
  workspace: any
}

export const TerraformWorkspace: FC<ITerraformWorkspace> = async ({
  orgId,
  installId,
  workspace,
}) => {
  const states = await getWorkspaceStates({
    orgId,
    workspaceId: workspace?.id,
  }).catch(console.error)

  const [currentRevision, resources, sandboxRuns, tokenRes] = await Promise.all(
    [
      getWorkspaceState({
        orgId,
        workspaceId: workspace?.id,
        stateId: states[0].id,
      }),
      getWorkspaceStateResources({
        workspaceId: workspace?.id,
        stateId: states[0]?.id,
        orgId,
      }),
      getInstallSandboxRuns({
        installId,
        orgId,
      }),
      getToken(),
    ]
  )

  const token = tokenRes.result.accessToken

  // const revisions = states.map((state, idx) => {
  //   const sandboxRun = sandboxRuns[idx]

  //   return [
  //     state.revision,
  //     <Time>{state.created_at}</Time>,
  //     `sandbox/${sandboxRun?.id}`,
  //   ]
  // })
  const revisions = states.map((state, idx) => {
    const sandboxRun = sandboxRuns[idx]

    return [
      state.revision,
      <Time key={idx}>{state.created_at}</Time>,
      `sandbox/${sandboxRun?.id}`,
    ]
  })

  const resourceList = resources.map((resource, idx) => {
    return [
      resource.type,
      resource.instances.length,
      resource.mode,
      resource.instances,
    ]
  })

  const outputs = currentRevision.data.outputs
  const outputList = Object.keys(outputs).map((key) => [
    key,
    outputs[key].type[0],
    JSON.stringify(outputs[key].value),
    '',
  ])

  return (
    <>
      <Section
        className="flex-initial"
        heading="Terraform state"
        childrenClassName="flex flex-col gap-4"
        actions={
          <WorkspaceManagementDropdown
            orgId={orgId}
            workspace={workspace}
            token={token}
          />
        }
      >
        <Tabs>
          <Tab title="Resources">
            <DataTable
              headers={['Type', 'Count', 'Mode', '']}
              initData={resourceList}
            />
          </Tab>
          <Tab title="Outputs">
            <DataTable
              headers={['Name', 'Type', 'Value', '']}
              initData={outputList}
            />
          </Tab>
          <Tab title="History">
            <DataTable
              headers={['Revision', 'Created at', '']}
              initData={revisions}
            />
          </Tab>
        </Tabs>
      </Section>
    </>
  )
}
