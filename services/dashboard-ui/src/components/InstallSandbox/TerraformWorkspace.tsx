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
  workspace: any
}

export const TerraformWorkspace: FC<ITerraformWorkspace> = async ({
  orgId,
  workspace,
}) => {
  const [states, tokenRes] = await Promise.all([
    getWorkspaceStates({
      orgId,
      workspaceId: workspace?.id,
    }).catch(console.error),
    getToken().catch(console.error),
  ])

  // Default to an "empty" message.
  // This will display if there are no revisions created in the workspace yet.
  let contents = (
    <Empty
      emptyTitle="No revisions yet"
      emptyMessage="The workspace has been created, but no state has been written."
      variant="history"
    />
  )

  if (states.length) {
    const [currentRevision, resources] = await Promise.all([
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
    ])

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

    const outputs = currentRevision.data.outputs || []
    const outputList = Object.keys(outputs).map((key, idx) => [
      <Text key={idx}>{key}</Text>,
      <Text key={idx}>{outputs[key].type[0]}</Text>,
      <Code key={idx}>{JSON.stringify(outputs[key].value)}</Code>,
    ])

    contents = (
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
    )
  }

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
            token={(tokenRes as any).result.accessToken}
          />
        }
      >
        {contents}
      </Section>
    </>
  )
}
