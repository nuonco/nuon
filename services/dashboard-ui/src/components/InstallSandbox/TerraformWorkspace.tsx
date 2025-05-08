import React, { FC } from 'react'
import {
  getWorkspaceStates,
  getWorkspaceState,
  getWorkspaceStateResources,
} from '@/lib'
import { Section } from '@/components/Card'
import { Empty } from '@/components/Empty'
import { Time } from '@/components/Time'
import { WorkspaceManagementDropdown } from '@/components/InstallSandbox/WorkspaceManagementDropdown'
import { DataTable } from '@/components/InstallSandbox/Table'
import { Tabs, Tab } from '@/components/InstallSandbox/Tabs'
import { Text, Code } from '@/components/Typography'
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

  const [currentRevision, resources, tokenRes] = await Promise.all([
    getWorkspaceState({
      orgId,
      workspaceId: workspace?.id,
      stateId: states[0]?.id,
    }).catch(console.error),
    getWorkspaceStateResources({
      workspaceId: workspace?.id,
      stateId: states[0]?.id,
      orgId,
    }).catch(console.error),
    getToken(),
  ])

  const token = tokenRes.result.accessToken

  if (states && resources && currentRevision) {
    const revisions = states.map((state: any, idx: number) => {
      return [
        <Text key={idx}>{state.revision}</Text>,
        <Time key={idx} time={state.created_at} />,
      ]
    })

    const resourceList = resources.map((resource, idx) => {
      if (resource.mode == 'managed') {
        return [
          <Text key={idx}>{resource.type}</Text>,
          <Text key={idx}>{resource.name}</Text>,
          <Text key={idx}>{resource.instances.length}</Text>,
        ]
      }
    })

    const datasourceList = resources.map((datasource, idx) => {
      if (datasource.mode == 'data') {
        return [
          <Text key={idx}>{datasource.type}</Text>,
          <Text key={idx}>{datasource.name}</Text>,
          <Text key={idx}>{datasource.instances.length}</Text>,
        ]
      }
    })

    const outputs = currentRevision.data.outputs
    const outputList = Object.keys(outputs).map((key, idx) => [
      <Text key={idx}>{key}</Text>,
      <Text key={idx}>{outputs[key].type[0]}</Text>,
      <Code key={idx}>{JSON.stringify(outputs[key].value)}</Code>,
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
                headers={['Type', 'Name', 'Count']}
                initData={resourceList}
              />
            </Tab>
            <Tab title="Data Sources">
              <DataTable
                headers={['Type', 'Name', 'Count']}
                initData={datasourceList}
              />
            </Tab>
            <Tab title="Outputs">
              <DataTable
                headers={['Name', 'Type', 'Value']}
                initData={outputList}
              />
            </Tab>
            <Tab title="History">
              <DataTable
                headers={['Revision', 'Created at']}
                initData={revisions}
              />
            </Tab>
          </Tabs>
        </Section>
      </>
    )
  } else {
    return null
  }
}
