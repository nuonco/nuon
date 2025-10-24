'use client'

import React, { type FC, useEffect, useRef } from 'react'
import { Funnel } from '@phosphor-icons/react'
import { ComponentConfigType } from './ComponentConfigType'
import { Button } from '@/components/old/Button'
import { Dropdown } from '@/components/old/Dropdown'
import { CheckboxInput } from '@/components/old/Input'

type TComponentConfigTypeText =
  | 'docker_build'
  | 'external_image'
  | 'helm_chart'
  | 'terraform_module'
  | 'kubernetes_manifest'
const FILTER_OPTIONS: Array<TComponentConfigTypeText> = [
  'docker_build',
  'external_image',
  'helm_chart',
  'terraform_module',
  'kubernetes_manifest',
]

const groupClasses = {
  docker_build: [
    'group/trace',
    'group-hover/trace:block group-hover/trace:opacity-100',
  ],
  external_image: [
    'group/debug',
    'group-hover/debug:block group-hover/debug:opacity-100',
  ],
  helm_chart: [
    'group/info',
    'group-hover/info:block group-hover/info:opacity-100',
  ],
  terraform_module: [
    'group/warn',
    'group-hover/warn:block group-hover/warn:opacity-100',
  ],
  kubernetes_manifest: [
    'group/warn',
    'group-hover/warn:block group-hover/warn:opacity-100',
  ],
}

interface IComponentTypeFilterDropdown {
  handleTypeFilter: any
  handleTypeOnlyFilter: any
  clearTypeFilter: any
  columnFilters: any
  isNotDropdown?: boolean
}

export const ComponentTypeFilterDropdown: FC<IComponentTypeFilterDropdown> = ({
  handleTypeFilter,
  handleTypeOnlyFilter,
  clearTypeFilter,
  columnFilters,
  isNotDropdown = false,
}) => {
  return isNotDropdown ? (
    <div className="w-full">
      <form>
        <div className="flex items-center gap-2">
          <Button
            className={`flex items-center justify-between w-fit mr-1.5 py-1 !px-1`}
            variant="ghost"
            type="button"
            onClick={clearTypeFilter}
          >
            <span className="flex items-center gap-1">
              <span className="font-semibold text-sm">Show all</span>
            </span>
          </Button>
          {FILTER_OPTIONS.map((opt) => (
            <div className="flex items-center gap-1" key={opt}>
              <CheckboxInput
                labelClassName="!w-auto !p-1.5  ml-1.5 rounded-md"
                name={opt}
                onChange={handleTypeFilter}
                checked={columnFilters?.at(0)?.value?.includes(opt)}
                value={opt}
              />
              <Button
                className={`flex items-center justify-between w-full mr-1.5 py-1 !px-1 ${groupClasses[opt].at(0)}`}
                variant="ghost"
                type="button"
                value={opt}
                onClick={
                  columnFilters?.at(0)?.value?.length === 1 &&
                  columnFilters?.at(0)?.value?.includes(opt)
                    ? clearTypeFilter
                    : handleTypeOnlyFilter
                }
              >
                <span className="flex items-center gap-1">
                  <span className="font-semibold text-sm">
                    <ComponentConfigType configType={opt} />
                  </span>
                </span>
                <span
                  className={`ml-2 text-sm self-end opacity-0 ${groupClasses[opt].at(1)} w-[40px]`}
                >
                  {columnFilters?.at(0)?.value?.length === 1 &&
                  columnFilters?.at(0)?.value?.includes(opt)
                    ? 'Reset'
                    : 'Only'}
                </span>
              </Button>
            </div>
          ))}
        </div>
      </form>
    </div>
  ) : (
    <Dropdown
      alignment="right"
      className="text-sm !font-medium !p-2 h-[32px]"
      id="logs-filter"
      text={
        <>
          <Funnel size="14" /> Filter ({columnFilters?.at(0)?.value?.length})
        </>
      }
      isDownIcon
    >
      <div className="min-w-[200px]">
        <form>
          <div className="py-2 flex flex-col gap-1">
            {FILTER_OPTIONS.map((opt) => (
              <div className="flex items-center gap-1" key={opt}>
                <CheckboxInput
                  labelClassName="!w-auto !p-1.5  ml-1.5 rounded-sm"
                  name={opt}
                  onChange={handleTypeFilter}
                  checked={columnFilters?.at(0)?.value?.includes(opt)}
                  value={opt}
                />
                <Button
                  className={`flex items-center justify-between w-full mr-1.5 py-1 !px-1 ${groupClasses[opt].at(0)}`}
                  variant="ghost"
                  type="button"
                  value={opt}
                  onClick={
                    columnFilters?.at(0)?.value?.length === 1 &&
                    columnFilters?.at(0)?.value?.includes(opt)
                      ? clearTypeFilter
                      : handleTypeOnlyFilter
                  }
                >
                  <span className="flex items-center gap-1">
                    <span className="font-semibold text-sm">
                      <ComponentConfigType configType={opt} />
                    </span>
                  </span>
                  <span
                    className={`text-sm self-end hidden ${groupClasses[opt].at(1)}`}
                  >
                    {columnFilters?.at(0)?.value?.length === 1 &&
                    columnFilters?.at(0)?.value?.includes(opt)
                      ? 'Reset'
                      : 'Only'}
                  </span>
                </Button>
              </div>
            ))}
          </div>

          <hr />
          <Button
            className="w-full !rounded-t-none !text-base flex items-center gap-2 pl-4"
            type="button"
            onClick={clearTypeFilter}
            variant="ghost"
          >
            Reset
          </Button>
        </form>
      </div>
    </Dropdown>
  )
}
