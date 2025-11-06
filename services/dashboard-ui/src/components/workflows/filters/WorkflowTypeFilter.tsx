'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { ChangeEvent, useCallback, useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { toSentenceCase } from '@/utils/string-utils'

const WORKFLOW_TYPE_OPTIONS = [
  'provision',
  'deprovision',
  'deprovision_sandbox',
  'manual_deploy',
  'input_update',
  'deploy_components',
  'teardown_component',
  'teardown_components',
  'reprovision_sandbox',
  'drift_run_reprovision_sandbox',
  'action_workflow_run',
  'sync_secrets',
  'drift_run',
  'reprovision',
]

type TTypeOption = (typeof WORKFLOW_TYPE_OPTIONS)[number]

const WORKFLOW_TYPE_PARAM = 'type'

export const WorkflowTypeFilter = () => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const workflowType = searchParams.get(WORKFLOW_TYPE_PARAM) || ''

  const updateTypeParam = useCallback(
    (type: string) => {
      const params = new URLSearchParams(searchParams.toString())

      if (type) {
        params.set(WORKFLOW_TYPE_PARAM, type)
      } else {
        params.delete(WORKFLOW_TYPE_PARAM)
      }

      router.replace(`?${params.toString()}`)
    },
    [router, searchParams]
  )

  const handleTypeChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      updateTypeParam(e.target.value)
    },
    [updateTypeParam]
  )

  const handleClearFilter = useCallback(() => {
    updateTypeParam('')
  }, [updateTypeParam])

  const typeOptions = useMemo(
    () =>
      WORKFLOW_TYPE_OPTIONS.map((type) => ({
        value: type,
        label: type,
      })),
    []
  )
  return (
    <Dropdown
      alignment="right"
      buttonClassName="!p-2"
      buttonText={
        <>
          <Icon variant="SlidersIcon" />
          Type filter
        </>
      }
      id="install-filter"
      variant="ghost"
    >
      <Menu className="!p-0 !w-68">
        <form onReset={handleClearFilter}>
          <div className="flex flex-col gap-0.5 max-h-[250px] overflow-y-auto w-full p-2 focus-visible:outline-1 focus-visible:outline-primary-600 rounded-md">
            {typeOptions.map(({ value, label }) => (
              <RadioInput
                key={value}
                checked={workflowType === value}
                labelProps={{
                  labelText: label,
                  labelTextProps: { family: 'mono', className: '!break-all' },
                }}
                name="workflow-type"
                onChange={handleTypeChange}
                value={value}
              />
            ))}
          </div>

          <div className="flex flex-col gap-0.5 px-2 pb-2 w-full">
            <hr />
            <Button className="mt-1" isMenuButton type="reset" variant="ghost">
              Clear
              <Icon variant="XIcon" />
            </Button>
          </div>
        </form>
      </Menu>
    </Dropdown>
  )
}
