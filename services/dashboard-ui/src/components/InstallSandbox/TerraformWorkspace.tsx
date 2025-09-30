import { SectionHeader } from '@/components/Card'
import { JsonView } from '@/components/Code'
import { Empty } from '@/components/Empty'
import { DataTable } from '@/components/InstallSandbox/Table'
import { Tabs, Tab } from '@/components/InstallSandbox/Tabs'
import { Text, Code } from '@/components/Typography'
import { getToken } from '@/components/admin-actions'
import {
  getTerraformState,
  getTerraformStates,
  getTerraformWorkspaceLock,
} from '@/lib'
import type { TTerraformState } from '@/types'
import { BackendModal } from './BackendModal'
import { UnlockModal } from './UnlockStateModal'

function getResourceAddresses(state: TTerraformState['values']['root_module']) {
  const addresses = []

  // Top-level resources
  if (state.resources && Array.isArray(state.resources)) {
    for (const res of state.resources) {
      if (res.address) addresses.push(res.address)
    }
  }

  // Resources in child_modules
  if (state.child_modules && Array.isArray(state.child_modules)) {
    for (const mod of state.child_modules) {
      if (mod.resources && Array.isArray(mod.resources)) {
        for (const res of mod.resources) {
          if (res.address) addresses.push(res.address)
        }
      }
    }
  }

  return addresses
}

export interface ITerraformWorkspace {
  orgId: string
  workspace: any
}

export const TerraformWorkspace = async ({
  orgId,
  workspace,
}: ITerraformWorkspace) => {
  const [{ data: states }, tokenRes, { data: lock }] = await Promise.all([
    getTerraformStates({
      orgId,
      workspaceId: workspace?.id,
    }),
    getToken().catch(console.error),
    getTerraformWorkspaceLock({
      orgId,
      workspaceId: workspace?.id,
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
    const { data: currentRevision } = await getTerraformState({
      orgId,
      workspaceId: workspace?.id,
      stateId: states?.at(0)?.id,
    })

    const resources = getResourceAddresses(currentRevision?.values?.root_module)

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
        <Code key={idx}>{outputs[key]?.value}</Code>
      ) : (
        <JsonView key={idx} data={outputs[key]?.value} />
      ),
    ])

    contents = (
      <>
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
      </>
    )
  }

  return (
    <>
      <SectionHeader
        heading="Terraform state"
        actions={
          <div className="flex items-center gap-4">
            {lock ? (
              <UnlockModal workspace={workspace} lock={lock} />
            ) : null}
            <BackendModal
              orgId={orgId}
              workspace={workspace}
              token={(tokenRes as any)?.result?.token}
            />
          </div>
        }
      />
      {contents}
    </>
  )
}
