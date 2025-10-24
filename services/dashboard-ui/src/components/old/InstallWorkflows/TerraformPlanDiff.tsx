'use client'

import React, { useState } from 'react'
import { TruncateValue } from './TerraformPlanTruncatedValue'

type ChangeAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'no-op'
  | 'create-before-destroy'
  | 'destroy-before-create'
  | 'replace'
  | 'read'

type TFResourceChange = {
  address: string
  mode: string
  type: string
  name: string
  change: {
    actions: ChangeAction[]
    before?: Record<string, any>
    after?: Record<string, any>
    after_unknown?: Record<string, any>
  }
}

type TFResourceDrift = {
  address: string
  mode: string
  type: string
  name: string
  change: {
    actions: ChangeAction[]
    before?: Record<string, any>
    after?: Record<string, any>
    after_unknown?: Record<string, any>
  }
}

function getActionColor(actions: ChangeAction[]): string {
  if (actions.length === 1 && actions[0] === 'read')
    return 'bg-blue-100 text-blue-700 border-blue-400 dark:bg-blue-500/10 dark:border-blue-600 dark:text-blue-200'
  if (actions.includes('create') && actions.includes('delete'))
    return 'bg-primary-100 text-primary-700 border-primary-400 dark:bg-primary-500/10 dark:border-primary-600 dark:text-primary-200'
  if (actions.includes('create'))
    return 'bg-green-100 text-green-700 border-green-400 dark:bg-green-500/10 dark:border-green-600 dark:text-green-200'
  if (actions.includes('update'))
    return 'bg-orange-100 text-orange-700 border-orange-400 dark:bg-orange-500/10 dark:border-orange-600 dark:text-orange-200'
  if (actions.includes('delete'))
    return 'bg-red-100 text-red-700 border-red-400 dark:bg-red-500/10 dark:border-red-600 dark:text-red-200'
  return 'bg-cool-grey-100 text-cool-grey-700 border-cool-grey-400 dark:bg-dark-grey-500/10 dark:border-cool-grey-500 dark:text-cool-grey-500'
}

function getDriftColor(): string {
  return 'bg-amber-100 text-amber-700 border-amber-400 dark:bg-amber-500/10 dark:border-amber-600 dark:text-amber-200'
}

function getActionLabel(actions: ChangeAction[]): string {
  if (actions.length === 1 && actions[0] === 'read') return 'Read'
  if (actions.includes('create') && actions.includes('delete')) return 'Replace'
  if (actions.includes('create')) return 'Create'
  if (actions.includes('update')) return 'Update'
  if (actions.includes('delete')) return 'Destroy'
  return 'No-op'
}

function diffFields(
  before: Record<string, any> = {},
  after: Record<string, any> = {}
) {
  const allKeys = Array.from(
    new Set([...Object.keys(before), ...Object.keys(after)])
  )
  return allKeys.map((key) => {
    if (before[key] !== after[key]) {
      return (
        <div
          className="flex gap-2 items-start my-1 overflow-x-scroll"
          key={key}
        >
          <span className="font-mono text-sm text-gray-400 dark:text-cool-grey-200 w-fit">
            {key}:
          </span>
          <span className="text-sm line-through text-red-600 bg-red-50 dark:text-red-50 dark:bg-red-600/10 px-1 rounded">
            <TruncateValue title={key}>
              {before[key] !== undefined
                ? JSON.stringify(before[key], null, 2)
                : ''}
            </TruncateValue>
          </span>
          <span className="text-sm text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 px-1 rounded">
            <TruncateValue title={key}>
              {after[key] !== undefined
                ? JSON.stringify(after[key], null, 2)
                : ''}
            </TruncateValue>
          </span>
        </div>
      )
    } else {
      return (
        <div className="flex gap-2 items-start my-1" key={key}>
          <span className="font-mono text-sm text-cool-grey-400 dark:text-cool-grey-200 w-fit">
            {key}:
          </span>
          <span className="text-sm text-cool-grey-700 dark:text-cool-grey-100 break-all">
            {before[key] === undefined || before[key] === 'undefined' ? (
              <i>Known after apply</i>
            ) : (
              <TruncateValue title={key}>
                {JSON.stringify(before[key])}
              </TruncateValue>
            )}
          </span>
        </div>
      )
    }
  })
}

function OutputChangesViewer({
  plan,
  hideNoOps,
}: {
  plan: any
  hideNoOps: boolean
}) {
  const [open, setOpen] = useState<Record<string, boolean>>({})

  function getOutputChangeActions(change: any): ChangeAction[] {
    if (Array.isArray(change?.actions)) {
      return change.actions as ChangeAction[]
    }
    return ['no-op']
  }

  const outputKeys = plan.output_changes ? Object.keys(plan.output_changes) : []
  const filteredKeys = hideNoOps
    ? outputKeys.filter((key) => {
        const acts = getOutputChangeActions(plan.output_changes[key])
        return !(acts.length === 1 && acts[0] === 'no-op')
      })
    : outputKeys

  if (!outputKeys.length) {
    return null
  }

  return (
    <div>
      <h3 className="font-bold text-base mb-2 mt-6">Output Changes</h3>
      <div className="w-full mx-auto space-y-2">
        {filteredKeys.length > 0 ? (
          filteredKeys.map((key) => {
            const change = plan.output_changes[key]
            const actions = getOutputChangeActions(change)
            const actionLabel = getActionLabel(actions)
            const color = getActionColor(actions)
            const isOpen = open[key] ?? false

            return (
              <div
                key={key}
                className={`border-l-4 shadow rounded ${color} relative transition-all`}
              >
                <button
                  onClick={() => setOpen((o) => ({ ...o, [key]: !isOpen }))}
                  className="w-full flex justify-between items-center px-4 py-3 gap-3 text-left focus:outline-none"
                >
                  <span className="font-mono text-sm font-semibold truncate">
                    {key}
                  </span>
                  <span
                    className={`px-2 py-1 rounded-full text-[11px] border ${color} !bg-white/50 dark:!bg-black/20`}
                  >
                    {actionLabel}
                  </span>
                </button>
                {isOpen && (
                  <div className="bg-cool-grey-50 dark:bg-dark-grey-200 px-6 py-4 border-t">
                    <div className="mb-4 text-sm text-cool-grey-600 dark:text-cool-grey-300">
                      <b>Output:</b> {key}
                    </div>
                    <div>
                      {(actions.includes('update') ||
                        actions.includes('create') ||
                        actions.includes('delete')) && (
                        <div className="my-2">
                          {diffFields(
                            { value: change.before },
                            { value: change.after }
                          )}
                        </div>
                      )}
                      {actions.includes('no-op') && (
                        <div className="my-2">
                          {diffFields(
                            { value: change.before },
                            { value: change.after }
                          )}
                        </div>
                      )}
                      {actions.length === 1 && actions[0] === 'read' && (
                        <div className="my-2 text-blue-700 dark:text-blue-200 text-sm">
                          Terraform will refresh this output value from the
                          provider.
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )
          })
        ) : (
          <div className="p-8 text-base text-center bg-black/5 text-cool-grey-800 dark:bg-white/5 dark:text-cool-grey-300 border rounded-md">
            No output changes in the Terraform plan.
          </div>
        )}
      </div>
    </div>
  )
}

function ResourceDriftViewer({
  resource_drift,
}: {
  resource_drift: TFResourceDrift[]
}) {
  const [open, setOpen] = useState<Record<string, boolean>>({})

  if (!resource_drift?.length) {
    return null
  }

  return (
    <div>
      <h3 className="font-bold text-base mb-2">Resource Drift</h3>
      <div className="w-full mx-auto space-y-2">
        {resource_drift.map((res: TFResourceDrift) => {
          const actionLabel = getActionLabel(res.change.actions)
          const color = getDriftColor()
          const isOpen = open[res.address] ?? false

          return (
            <div
              key={res.address}
              className={`border-l-4 shadow rounded ${color} relative transition-all`}
            >
              <button
                onClick={() =>
                  setOpen((o) => ({ ...o, [res.address]: !isOpen }))
                }
                className="w-full flex justify-between items-center px-4 py-3 gap-3 text-left focus:outline-none"
              >
                <span className="font-mono text-sm font-semibold truncate">
                  {res.address}
                </span>
                <div className="flex items-center gap-2">
                  <span className="text-xs bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-200 px-2 py-1 rounded-full border border-amber-300 dark:border-amber-600">
                    DRIFT
                  </span>
                  <span
                    className={`px-2 py-1 rounded-full text-[11px] border ${color} !bg-white/50 dark:!bg-black/20`}
                  >
                    {actionLabel}
                  </span>
                </div>
              </button>
              {isOpen && (
                <div className="bg-cool-grey-50 dark:bg-dark-grey-200 px-6 py-4 border-t">
                  <div className="mb-4 text-sm text-cool-grey-600 dark:text-cool-grey-300">
                    <b>Type:</b> {res.type} &nbsp;
                    <b>Name:</b> {res.name}
                  </div>
                  <div className="mb-2 text-amber-700 dark:text-amber-200 text-sm bg-amber-50 dark:bg-amber-900/20 p-2 rounded border border-amber-200 dark:border-amber-700">
                    ⚠️ This resource has drifted from its expected state.
                  </div>
                  <div>
                    {res.change.actions.includes('create') &&
                      !res.change.before && (
                        <div className="text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 rounded p-2 text-[11px] overflow-x-auto">
                          {Object.keys(res.change.after).map((k) => (
                            <span className="flex items-start gap-2" key={k}>
                              <span>{k}:</span>
                              <TruncateValue title={k}>
                                {res.change.after[k] || 'null'}
                              </TruncateValue>
                            </span>
                          ))}
                        </div>
                      )}
                    {res.change.actions.includes('delete') &&
                      !res.change.after && (
                        <div className="text-red-700 bg-red-50 dark:text-red-50 dark:bg-red-600/10 rounded p-2 text-[11px] overflow-x-auto">
                          {Object.keys(res.change.before).map((k) => (
                            <span className="flex items-start gap-2" key={k}>
                              <span>{k}:</span>
                              <TruncateValue title={k}>
                                {res.change.before[k] || 'null'}
                              </TruncateValue>
                            </span>
                          ))}
                        </div>
                      )}
                    {res.change.actions.includes('update') && (
                      <div className="my-2">
                        {diffFields(res.change.before, res.change.after)}
                      </div>
                    )}
                    {res.change.actions.includes('create') &&
                      res.change.actions.includes('delete') && (
                        <div className="my-2">
                          {diffFields(res.change.before, res.change.after)}
                        </div>
                      )}
                    {res.change.actions.length === 1 &&
                      res.change.actions[0] === 'read' && (
                        <div className="my-2 text-blue-700 dark:text-blue-200 text-sm">
                          Terraform will refresh this resource from the
                          provider.
                        </div>
                      )}
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function ResourceChangesViewer({
  resource_changes,
  hideNoOps,
}: {
  resource_changes: TFResourceChange[]
  hideNoOps: boolean
}) {
  const [open, setOpen] = useState<Record<string, boolean>>({})

  const displayedResources = hideNoOps
    ? resource_changes.filter(
        (res: TFResourceChange) =>
          !(
            res.change.actions.length === 1 && res.change.actions[0] === 'no-op'
          )
      )
    : resource_changes

  if (!resource_changes.length) {
    return null
  }

  return (
    <div>
      <h3 className="font-bold text-base mb-2 mt-6">Resource Changes</h3>
      <div className="w-full mx-auto space-y-2">
        {displayedResources.length > 0 ? (
          displayedResources.map((res: TFResourceChange) => {
            const actionLabel = getActionLabel(res.change.actions)
            const color = getActionColor(res.change.actions)
            const isOpen = open[res.address] ?? false
            return (
              <div
                key={res.address}
                className={`border-l-4 shadow rounded ${color} relative transition-all`}
              >
                <button
                  onClick={() =>
                    setOpen((o) => ({ ...o, [res.address]: !isOpen }))
                  }
                  className="w-full flex justify-between items-center px-4 py-3 gap-3 text-left focus:outline-none"
                >
                  <span className="font-mono text-sm font-semibold truncate">
                    {res.address}
                  </span>
                  <span
                    className={`px-2 py-1 rounded-full text-[11px] border ${color} !bg-white/50 dark:!bg-black/20`}
                  >
                    {actionLabel}
                  </span>
                </button>
                {isOpen && (
                  <div className="bg-cool-grey-50 dark:bg-dark-grey-200 px-6 py-4 border-t">
                    <div className="mb-4 text-sm text-cool-grey-600 dark:text-cool-grey-300">
                      <b>Type:</b> {res.type} &nbsp;
                      <b>Name:</b> {res.name}
                    </div>
                    <div>
                      {res.change.actions.includes('create') &&
                        !res.change.before && (
                          <div className="text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 rounded p-2 text-[11px] overflow-x-auto">
                            {Object.keys(res.change.after).map((k) => (
                              <span className="flex items-start gap-2" key={k}>
                                <span>{k}:</span>
                                <TruncateValue title={k}>
                                  {res.change.after[k] || 'null'}
                                </TruncateValue>
                              </span>
                            ))}
                          </div>
                        )}
                      {res.change.actions.includes('delete') &&
                        !res.change.after && (
                          <div className="text-red-700 bg-red-50 dark:text-red-50 dark:bg-red-600/10 rounded p-2 text-[11px] overflow-x-auto">
                            {Object.keys(res.change.before).map((k) => (
                              <span className="flex items-start gap-2" key={k}>
                                <span>{k}:</span>
                                <TruncateValue title={k}>
                                  {res.change.before[k] || 'null'}
                                </TruncateValue>
                              </span>
                            ))}
                          </div>
                        )}
                      {res.change.actions.includes('update') && (
                        <div className="my-2">
                          {diffFields(res.change.before, res.change.after)}
                        </div>
                      )}
                      {res.change.actions.includes('create') &&
                        res.change.actions.includes('delete') && (
                          <div className="my-2">
                            {diffFields(res.change.before, res.change.after)}
                          </div>
                        )}
                      {res.change.actions.length === 1 &&
                        res.change.actions[0] === 'read' && (
                          <div className="my-2 text-blue-700 dark:text-blue-200 text-sm">
                            Terraform will refresh this resource from the
                            provider.
                          </div>
                        )}
                    </div>
                  </div>
                )}
              </div>
            )
          })
        ) : (
          <div className="p-8 text-base text-center bg-black/5 text-cool-grey-800 dark:bg-white/5 dark:text-cool-grey-300 border rounded-md">
            No resource changes in the Terraform plan.
          </div>
        )}
      </div>
    </div>
  )
}

export function TerraformPlanViewer({ plan, showNoops = false }: { plan: any, showNoops?: boolean }) {
  // Default: hide no-op changes, so set true
  const [hideNoOps, setHideNoOps] = useState<boolean>(!showNoops)

  const hasResourceDrift =
    Array.isArray(plan?.resource_drift) && plan.resource_drift.length > 0
  const hasResourceChanges =
    Array.isArray(plan?.resource_changes) && plan.resource_changes.length > 0
  const hasOutputChanges =
    plan?.output_changes && Object.keys(plan.output_changes).length > 0

  if (!hasResourceDrift && !hasResourceChanges && !hasOutputChanges) {
    return (
      <div>
        <div className="flex items-center mb-4">
          <input
            id="show-noops"
            type="checkbox"
            className="mr-2"
            checked={!hideNoOps}
            onChange={() => setHideNoOps((v) => !v)}
          />
          <label htmlFor="show-noops" className="text-sm">
            Show no-op changes
          </label>
        </div>
        <div className="p-8 text-base text-center bg-black/5 text-cool-grey-800 dark:bg-white/5 dark:text-cool-grey-300 border rounded-md">
          No changes found in the Terraform plan.
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center mb-4">
        <input
          id="show-noops"
          type="checkbox"
          className="mr-2"
          checked={!hideNoOps}
          onChange={() => setHideNoOps((v) => !v)}
        />
        <label htmlFor="show-noops" className="text-sm">
          Show no-op changes
        </label>
      </div>
      {hasResourceDrift && (
        <ResourceDriftViewer resource_drift={plan.resource_drift} />
      )}
      {hasResourceChanges && (
        <ResourceChangesViewer
          resource_changes={plan.resource_changes}
          hideNoOps={hideNoOps}
        />
      )}
      {hasOutputChanges && (
        <OutputChangesViewer plan={plan} hideNoOps={hideNoOps} />
      )}
    </div>
  )
}
