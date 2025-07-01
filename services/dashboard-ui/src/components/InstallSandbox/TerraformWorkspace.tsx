import React, { FC } from 'react'
import {
  getWorkspaceStates,
  getWorkspaceState,
  getWorkspaceStateResources,
} from '@/lib'
import { SectionHeader } from '@/components/Card'
import { JsonView } from '@/components/Code'
import { Empty } from '@/components/Empty'
import { Time } from '@/components/Time'
import { DataTable } from '@/components/InstallSandbox/Table'
import { Tabs, Tab } from '@/components/InstallSandbox/Tabs'
import { Text, Code } from '@/components/Typography'
import { getToken } from '@/components/admin-actions'
import { BackendModal } from './BackendModal'
import { UnlockModal } from './UnlockStateModal'
import { nueQueryData } from '@/utils'

export interface ITerraformWorkspace {
  orgId: string
  workspace: any
}

export const TerraformWorkspace: FC<ITerraformWorkspace> = async ({
  orgId,
  workspace,
}) => {
  const [states, tokenRes, lockRes] = await Promise.all([
    getWorkspaceStates({
      orgId,
      workspaceId: workspace?.id,
    }).catch(console.error),
    getToken().catch(console.error),
    nueQueryData({
      orgId,
      path: `terraform-workspaces/${workspace?.id}/lock`,
    }),
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
        stateId: states.at(-1)?.id,
      }).catch(console.error),
      getWorkspaceStateResources({
        workspaceId: workspace?.id,
        stateId: states[0]?.id,
        orgId,
      }).catch(console.error),
    ])

    const revisions = states?.map((state: any, idx: number) => {
      return [
        <Text key={idx}>{state.revision}</Text>,
        <Time key={idx} time={state.created_at} />,
      ]
    })

    const resourceList = resources?.map((resource, idx) => {
      return [
        <span className="block overflow-hidden truncate" key={idx}>
          <Text
            className="text-ellipsis overflow-hidden py-2 w-full"
            variant="mono-12"
            style={{ display: 'inline' }}
          >
            {resource}
          </Text>
        </span>,
      ]
    })

    const datasourceList = resources?.map((datasource, idx) => {
      if (datasource.mode == 'data') {
        return [
          <Text key={idx}>{datasource.type}</Text>,
          <Text key={idx}>{datasource.name}</Text>,
          <Text key={idx}>{datasource.instances.length}</Text>,
        ]
      }
    })

    const outputs = currentRevision?.values?.outputs || []
    const outputList = Object.keys(outputs).map((key, idx) => [
      <span key={idx} className="flex flex-col">
        <Text variant="med-12">{key}</Text>
        <Text variant="mono-12">
          {Array.isArray(outputs[key].type)
            ? outputs[key].type[0]
            : outputs[key].type}
        </Text>
      </span>,
      outputs[key]?.type === 'string' || outputs[key]?.type === 'number' ? (
        <Code>{outputs[key]?.value}</Code>
      ) : (
        <JsonView key={idx} data={outputs[key]?.value} />
      ),
    ])

    contents = (
      <Tabs>
        {resourceList && (
          <Tab title="Resources list">
            <div className="flex flex-col gap-4">
              <Text>State addresses</Text>

              <div className="flex flex-col divide-y">{resourceList}</div>
            </div>
          </Tab>
        )}

        {outputList && (
          <Tab title="Outputs">
            <DataTable
              headers={['Name & Type', 'Value']}
              initData={outputList}
            />
          </Tab>
        )}
      </Tabs>
    )
  }

  return (
    <>
      <SectionHeader
        heading="Terraform state"
        actions={
          <div className="flex items-center gap-4">
            {lockRes?.data ? (
              <UnlockModal
                workspace={workspace}
                orgId={orgId}
                lock={lockRes?.data}
              />
            ) : null}
            <BackendModal
              orgId={orgId}
              workspace={workspace}
              token={(tokenRes as any)?.result?.accessToken}
            />
          </div>
        }
      />
      {contents}
    </>
  )
}
