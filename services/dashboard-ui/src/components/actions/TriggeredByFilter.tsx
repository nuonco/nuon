'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { ChangeEvent, useCallback, useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { toSentenceCase } from '@/utils/string-utils'

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

type TTriggerOption = (typeof TRIGGER_OPTIONS)[number]

const TRIGGER_PARAM = 'trigger_types'

export const TriggeredByFilter = () => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const triggerType = searchParams.get(TRIGGER_PARAM) || ''

  const updateTriggerParam = useCallback(
    (type: string) => {
      const params = new URLSearchParams(searchParams.toString())

      if (type) {
        params.set(TRIGGER_PARAM, type)
      } else {
        params.delete(TRIGGER_PARAM)
      }

      router.replace(`?${params.toString()}`)
    },
    [router, searchParams]
  )

  const handleTriggerChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      updateTriggerParam(e.target.value)
    },
    [updateTriggerParam]
  )

  const handleClearFilter = useCallback(() => {
    updateTriggerParam('')
  }, [updateTriggerParam])

  const triggerOptions = useMemo(
    () =>
      TRIGGER_OPTIONS.map((trigger) => ({
        value: trigger,
        label: toSentenceCase(trigger),
      })),
    []
  )

  return (
    <Dropdown
      alignment="right"
      className="text-sm !font-medium !p-2 h-[32px]"
      buttonText={
        <>
          <Icon variant="SlidersIcon" />
          Trigger filter
        </>
      }
      id="install-filter"
      variant="ghost"
      isOpen
    >
      <Menu className="!p-0">
        <form onReset={handleClearFilter}>
          <div className="flex flex-col gap-0.5 max-h-[250px] overflow-y-auto w-full p-2 focus-visible:outline-1 focus-visible:outline-primary-600 rounded-md">
            {triggerOptions.map(({ value, label }) => (
              <RadioInput
                key={value}
                checked={triggerType === value}
                labelProps={{
                  labelText: label,
                }}
                name="triggered-by-type"
                onChange={handleTriggerChange}
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
