'use client'

import React, { type FC } from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Dropdown } from '@/components/Dropdown'
import { RadioInput } from '@/components/Input'

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

export interface IInstallActionTriggerFilter {
  handleTriggerFilter: any
  clearTriggerFilter: any
}

export const InstallActionTriggerFilter: FC<IInstallActionTriggerFilter> = ({
  handleTriggerFilter,
  clearTriggerFilter,
}) => {
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
            {TRIGGER_OPTIONS?.map((t) => (
              <RadioInput
                key={t}
                labelClassName="font-mono"
                name="trigger-filter"
                onChange={handleTriggerFilter}
                value={t}
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
