'use client'

import React from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Dropdown } from '@/components/old/Dropdown'
import { RadioInput } from '@/components/old/Input'
import { useRouter, useSearchParams } from 'next/navigation'

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
]

export const InstallActionTriggerFilter: React.FC = () => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const triggerType = searchParams.get('trigger_types') || ''

  // Select trigger type in URL
  const setTriggerTypeInUrl = (type: string) => {
    const params = new URLSearchParams(searchParams.toString())
    if (type) {
      params.set('trigger_types', type)
    } else {
      params.delete('trigger_types')
    }
    router.replace(`?${params.toString()}`)
  }

  const handleTriggerFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTriggerTypeInUrl(e.target.value)
  }

  const clearTriggerFilter = () => {
    setTriggerTypeInUrl('')
  }

  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      id="install-filter"
      text="Recent trigger"
      alignment="right"
    >
      <div>
        <form>
          <div className="min-w-[250px] max-h-[350px] overflow-y-auto">
            {TRIGGER_OPTIONS.map((t) => (
              <RadioInput
                key={t}
                labelClassName="font-mono"
                name="trigger-filter"
                onChange={handleTriggerFilter}
                value={t}
                checked={triggerType === t}
                labelText={t}
              />
            ))}
          </div>
          <hr />
          <Button
            className="w-full !rounded-t-none !text-sm flex items-center gap-2 pl-4 py-3"
            type="reset"
            onClick={clearTriggerFilter}
            variant="ghost"
          >
            <X />
            Clear
          </Button>
        </form>
      </div>
    </Dropdown>
  )
}
