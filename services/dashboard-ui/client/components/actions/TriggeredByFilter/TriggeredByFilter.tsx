import type { ChangeEvent } from 'react'
import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

const TRIGGER_OPTIONS = [
  'manual',
  'cron',
  'pre-deploy-component',
  'post-deploy-component',
  'pre-teardown-component',
  'post-teardown-component',
  'pre-secrets-sync',
  'post-secrets-sync',
  'pre-provision',
  'post-provision',
  'pre-reprovision',
  'post-reprovision',
  'pre-deprovision',
  'post-deprovision',
  'pre-deploy-all-components',
  'post-deploy-all-components',
  'pre-teardown-all-components',
  'post-teardown-all-components',
  'pre-deprovision-sandbox',
  'post-deprovision-sandbox',
  'pre-reprovision-sandbox',
  'post-reprovision-sandbox',
  'pre-update-inputs',
  'post-update-inputs',
] as const

interface ITriggeredByFilter {
  triggerType: string
  onTriggerChange: (e: ChangeEvent<HTMLInputElement>) => void
  onClearFilter: () => void
}

export const TriggeredByFilter = ({
  triggerType,
  onTriggerChange,
  onClearFilter,
}: ITriggeredByFilter) => {
  const triggerOptions = useMemo(
    () =>
      TRIGGER_OPTIONS.map((trigger) => ({
        value: trigger,
        label: trigger,
      })),
    []
  )

  return (
    <Dropdown
      alignment="right"
      buttonText={
        <>
          <Icon variant="SlidersIcon" />
          Trigger filter
        </>
      }
      id="install-filter"
    >
      <Menu className="!p-0 !w-68">
        <form onReset={onClearFilter}>
          <div className="flex flex-col gap-0.5 max-h-[250px] overflow-y-auto w-full p-2 focus-visible:outline-1 focus-visible:outline-primary-600 rounded-md">
            {triggerOptions.map(({ value, label }) => (
              <RadioInput
                key={value}
                checked={triggerType === value}
                labelProps={{
                  labelText: label,
                  labelTextProps: { family: 'mono' },
                }}
                name="triggered-by-type"
                onChange={onTriggerChange}
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
