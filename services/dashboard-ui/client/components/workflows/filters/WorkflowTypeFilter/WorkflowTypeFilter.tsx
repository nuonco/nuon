import { type ChangeEvent, useCallback, useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

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
] as const

interface IWorkflowTypeFilter {
  workflowType: string
  onTypeChange: (type: string) => void
  onClearFilter: () => void
}

export const WorkflowTypeFilter = ({
  workflowType,
  onTypeChange,
  onClearFilter,
}: IWorkflowTypeFilter) => {
  const handleTypeChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => onTypeChange(e.target.value),
    [onTypeChange]
  )

  const typeOptions = useMemo(
    () => WORKFLOW_TYPE_OPTIONS.map((type) => ({ value: type, label: type })),
    []
  )

  return (
    <Dropdown
      alignment="right"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" />
          Type
        </>
      }
      id="workflow-type-filter"
    >
      <Menu className="!p-0 !w-68">
        <form onReset={onClearFilter}>
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
