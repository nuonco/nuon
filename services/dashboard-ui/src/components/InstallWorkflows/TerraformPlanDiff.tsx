'use client'

import React, { useState } from 'react'
import { JsonView } from '@/components/Code'

// Original types for classic plan
type ChangeAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'no-op'
  | 'create-before-destroy'
  | 'destroy-before-create'
  | 'replace'

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

type TerraformPlanClassic = {
  resource_changes: TFResourceChange[]
}

type PlanFormat = 'resource_changes' | 'state_plan' | 'unknown'

function detectPlanFormat(plan: any): PlanFormat {
  if (Array.isArray(plan?.resource_changes)) return 'resource_changes'
  if (plan?.configuration && plan?.planned_values && plan?.output_changes)
    return 'state_plan'
  return 'unknown'
}

function getActionColor(actions: ChangeAction[]): string {
  if (actions.includes('create') && actions.includes('delete'))
    return 'bg-primary-100 text-primary-700 border-primary-400 dark:bg-primary-500/10 dark:border-primary-600 dark:text-primary-200' // replace
  if (actions.includes('create'))
    return 'bg-green-100 text-green-700 border-green-400 dark:bg-green-500/10 dark:border-green-600 dark:text-green-200'
  if (actions.includes('update'))
    return 'bg-orange-100 text-orange-700 border-orange-400 dark:bg-orange-500/10 dark:border-orange-600 dark:text-orange-200'
  if (actions.includes('delete'))
    return 'bg-red-100 text-red-700 border-red-400 dark:bg-red-500/10 dark:border-red-600 dark:text-red-200'
  return 'bg-cool-grey-100 text-cool-grey-700 border-cool-grey-400 dark:bg-dark-grey-500/10 dark:border-cool-grey-500 dark:text-cool-grey-500'
}

function getActionLabel(actions: ChangeAction[]): string {
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
            {before[key] !== undefined
              ? JSON.stringify(before[key], null, 2)
              : ''}
          </span>
          <span className="text-sm text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 px-1 rounded">
            {after[key] !== undefined
              ? JSON.stringify(after[key], null, 2)
              : ''}
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
              JSON.stringify(before[key])
            )}
          </span>
        </div>
      )
    }
  })
}

// ---- StatePlanViewer styled like the resource_changes viewer ----
function StatePlanViewer({ plan }: { plan: any }) {
  const [open, setOpen] = useState<Record<string, boolean>>({})

  // Helper for state plan "output_changes" -- convert to ChangeAction[]
  function getOutputChangeActions(change: any): ChangeAction[] {
    // Usually it's an array like ["no-op"], ["create"], etc.
    if (Array.isArray(change?.actions)) {
      return change.actions as ChangeAction[]
    }
    return ['no-op']
  }

  // Render output changes as cards
  const outputKeys = plan.output_changes ? Object.keys(plan.output_changes) : []

  return (
    <div className="w-full mx-auto space-y-2">
      {outputKeys.length > 0 ? (
        outputKeys.map((key) => {
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
                  </div>
                  {/* <div className="mt-2 text-sm text-gray-400">
                      <span className="flex gap-2">
                      <b>Unknown after: </b> {String(change.after_unknown)}
                      </span>

                      <span className="flex gap-2">
                      <b>Sensitive after: </b> {String(change.after_sensitive)}
                      </span>

                      <span className="flex gap-2">
                      <b>Sensitive before: </b>{' '}
                      {String(change.before_sensitive)}
                      </span>
                      </div> */}
                </div>
              )}
            </div>
          )
        })
      ) : (
        <div className="p-8 text-base text-center text-gray-500">
          No output changes in the Terraform plan.
        </div>
      )}

      {/* Planned Outputs */}
      <div className="border-l-4 shadow rounded bg-cool-grey-100 dark:bg-dark-grey-500/10 mt-4">
        <button
          onClick={() =>
            setOpen((o) => ({ ...o, planned_outputs: !o.planned_outputs }))
          }
          className="w-full flex justify-between items-center px-4 py-3 gap-3 text-left focus:outline-none"
        >
          <span className="font-mono text-sm font-semibold truncate">
            Planned Outputs
          </span>
        </button>
        {open.planned_outputs && (
          <div className="bg-cool-grey-50 dark:bg-dark-grey-200 px-6 py-4 border-t">
            <JsonView data={plan.planned_values?.outputs || {}} />
          </div>
        )}
      </div>

      {/* Configuration Resources */}
      <div className="border-l-4 shadow rounded bg-cool-grey-100 dark:bg-dark-grey-500/10 mt-4">
        <button
          onClick={() =>
            setOpen((o) => ({
              ...o,
              configuration_resources: !o.configuration_resources,
            }))
          }
          className="w-full flex justify-between items-center px-4 py-3 gap-3 text-left focus:outline-none"
        >
          <span className="font-mono text-sm font-semibold truncate">
            Configuration Resources
          </span>
        </button>
        {open.configuration_resources && (
          <div className="bg-cool-grey-50 dark:bg-dark-grey-200 px-6 py-4 border-t">
            <JsonView data={plan.configuration?.root_module?.resources || []} />
          </div>
        )}
      </div>
    </div>
  )
}

export function TerraformPlanViewer({ plan }: { plan: any }) {
  const [open, setOpen] = useState<Record<string, boolean>>({})
  const format = detectPlanFormat(plan)

  // Classic format
  if (format === 'resource_changes') {
    if (!plan?.resource_changes?.length) {
      return plan ? (
        <JsonView data={plan} />
      ) : (
        <div className="p-8 text-base text-center">
          No changes found in the Terraform plan.
        </div>
      )
    }

    return (
      <div className="w-full mx-auto space-y-2">
        <div className="text-sm text-gray-400 mb-2">
          Detected resource changes.
        </div>
        {plan.resource_changes.map((res: TFResourceChange) => {
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
                        <pre className="text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 rounded p-2 text-[11px] overflow-x-auto">
                          {JSON.stringify(res.change.after, null, 2)}
                        </pre>
                      )}
                    {res.change.actions.includes('delete') &&
                      !res.change.after && (
                        <pre className="text-red-700 bg-red-50 dark:text-red-50 dark:bg-red-600/10 rounded p-2 text-[11px] overflow-x-auto">
                          {JSON.stringify(res.change.before, null, 2)}
                        </pre>
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
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    )
  }

  // New state file style
  if (format === 'state_plan') {
    return (
      <div>
        <div className="text-sm text-gray-400 mb-2">
          Detected output changes.
        </div>
        <StatePlanViewer plan={plan} />
      </div>
    )
  }

  // fallback
  return (
    <div>
      <div className="text-sm text-gray-400 mb-2">
        Unknown Terraform plan format. Showing as JSON.
      </div>
      <JsonView data={plan} />
    </div>
  )
}
