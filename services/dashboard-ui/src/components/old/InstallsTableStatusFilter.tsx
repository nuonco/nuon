'use client'

import React, { type FC } from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Dropdown } from '@/components/old/Dropdown'
import { RadioInput } from '@/components/old/Input'

export interface IInstallsTableStatusFilter {
  handleStatusFilter: any
  clearStatusFilter: any
}

export const InstallsTableStatusFilter: FC<IInstallsTableStatusFilter> = ({
  handleStatusFilter,
  clearStatusFilter,
}) => {
  return (
    <Dropdown
      className="text-sm !font-medium !p-2 h-[32px]"
      id="install-filter"
      text="Filter"
      alignment="right"
    >
      <div>
        <form>
          <RadioInput
            name="status-filter"
            onChange={handleStatusFilter}
            value="error"
            labelText="Error"
          />

          <RadioInput
            name="status-filter"
            onChange={handleStatusFilter}
            value="processing"
            labelText="Processing"
          />

          <RadioInput
            name="status-filter"
            onChange={handleStatusFilter}
            value="noop"
            labelText="NOOP"
          />

          <RadioInput
            name="status-filter"
            onChange={handleStatusFilter}
            value="active"
            labelText="Active"
          />
          <hr />
          <Button
            className="w-full !rounded-t-none !text-sm flex items-center gap-2 pl-4"
            type="reset"
            onClick={clearStatusFilter}
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
