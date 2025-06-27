'use client'

import React, { useEffect, useState } from 'react'
import { JsonView, CodeViewer } from '@/components/Code'

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

type TerraformPlan = {
  resource_changes: TFResourceChange[]
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
          <span className="font-mono text-[11px] text-gray-400 dark:text-cool-grey-200 w-fit">
            {key}:
          </span>
          <span className="text-[11px] line-through text-red-600 bg-red-50 dark:text-red-50 dark:bg-red-600/10 px-1 rounded">
            {before[key] !== undefined
              ? JSON.stringify(before[key], null, 2)
              : ''}
          </span>
          <span className="text-[11px] text-green-700 bg-green-50 dark:text-green-50 dark:bg-green-600/10 px-1 rounded">
            {after[key] !== undefined
              ? JSON.stringify(after[key], null, 2)
              : ''}
          </span>
        </div>
      )
    } else {
      return (
        <div className="flex gap-2 items-start my-1" key={key}>
          <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-200 w-fit">
            {key}:
          </span>
          <span className="text-[11px] text-cool-grey-700 dark;text-cool-grey-100 break-all">
            {JSON.stringify(before[key])}
          </span>
        </div>
      )
    }
  })
}

export function TerraformPlanViewer({ plan }: { plan: TerraformPlan }) {
  const [open, setOpen] = useState<Record<string, boolean>>({})

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
      {plan.resource_changes.map((res) => {
        const actionLabel = getActionLabel(res.change.actions)
        const color = getActionColor(res.change.actions)
        const isOpen = open[res.address] ?? false
        return (
          <div
            key={res.address}
            className={`border-l-4 shadow rounded ${color} relative transition-all`}
          >
            <button
              onClick={() => setOpen((o) => ({ ...o, [res.address]: !isOpen }))}
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
