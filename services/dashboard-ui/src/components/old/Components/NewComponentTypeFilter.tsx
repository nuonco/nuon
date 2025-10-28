'use client'

import React from 'react'
import { Funnel } from '@phosphor-icons/react'
import { useRouter, useSearchParams } from 'next/navigation'
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
    'group/error',
    'group-hover/error:block group-hover/error:opacity-100',
  ],
}

interface IComponentTypeFilterDropdown {
  isNotDropdown?: boolean
}

export const ComponentTypeFilterDropdown: React.FC<
  IComponentTypeFilterDropdown
> = ({ isNotDropdown = false }) => {
  const router = useRouter()
  const searchParams = useSearchParams()

  // Parse types from search param
  const typesParam = searchParams.get('types')
  // If types is missing or empty, treat all as selected
  const allSelected = !typesParam || typesParam === ''
  const selectedTypes: TComponentConfigTypeText[] = allSelected
    ? FILTER_OPTIONS
    : typesParam!
        .split(',')
        .filter((v): v is TComponentConfigTypeText =>
          FILTER_OPTIONS.includes(v as any)
        )

  const setTypesInUrl = (types: TComponentConfigTypeText[]) => {
    // If this code might run during SSR, bail out since it relies on window/router
    if (typeof window === 'undefined') return

    const params = new URLSearchParams(window.location.search)

    // Normalize existing and new types to strings for reliable comparison
    const existingTypes = params.get('types')
      ? params.get('types')!.split(',')
      : []
    const newTypes = types.map(String)

    // Compare as sets (order doesn't matter)
    const setsEqual = (a: string[], b: string[]) => {
      if (a.length !== b.length) return false
      const s = new Set(a)
      return b.every((x) => s.has(x))
    }

    // Only reset pagination/offset when the selected types actually change
    if (!setsEqual(existingTypes, newTypes)) {
      params.delete('offset')
    }

    // If all types are selected, remove the param (API default)
    if (newTypes.length === FILTER_OPTIONS.length) {
      params.delete('types')
    } else if (newTypes.length > 0) {
      params.set('types', newTypes.join(','))
    } else {
      params.delete('types')
    }

    const query = params.toString()
    router.replace(`${window.location.pathname}${query ? `?${query}` : ''}`)
  }

  // Checkbox toggle handler
  const handleTypeFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value as TComponentConfigTypeText
    let newTypes: TComponentConfigTypeText[]
    if (e.target.checked) {
      // Add type if not present
      newTypes = Array.from(new Set([...selectedTypes, value]))
    } else {
      // Remove type
      newTypes = selectedTypes.filter((t) => t !== value)
    }
    setTypesInUrl(newTypes)
  }

  // "Only" button handler
  const handleTypeOnlyFilter = (e: React.MouseEvent<HTMLButtonElement>) => {
    const value = e.currentTarget.value as TComponentConfigTypeText
    setTypesInUrl([value])
  }

  // Show all: check all, remove param
  const handleShowAll = () => setTypesInUrl(FILTER_OPTIONS)

  const renderFilters = (buttonLabelClass: string, onlyLabelClass: string) =>
    FILTER_OPTIONS.map((opt) => (
      <div className="flex items-center gap-1" key={opt}>
        <CheckboxInput
          labelClassName="!w-auto !p-1.5 ml-1.5 rounded-md"
          name={opt}
          onChange={handleTypeFilter}
          checked={selectedTypes.includes(opt)}
          value={opt}
        />
        <Button
          className={`flex items-center justify-between w-full mr-1.5 py-1 !px-1 ${groupClasses[opt][0]} ${buttonLabelClass}`}
          variant="ghost"
          type="button"
          value={opt}
          onClick={
            selectedTypes.length === 1 && selectedTypes.includes(opt)
              ? handleShowAll
              : handleTypeOnlyFilter
          }
        >
          <span className="flex items-center gap-1">
            <span className="font-semibold text-xs">
              <ComponentConfigType configType={opt} />
            </span>
          </span>
          <span
            className={`ml-2 text-xs self-end opacity-0 ${groupClasses[opt][1]} ${onlyLabelClass}`}
          >
            {selectedTypes.length === 1 && selectedTypes.includes(opt)
              ? 'Reset'
              : 'Only'}
          </span>
        </Button>
      </div>
    ))

  return isNotDropdown ? (
    <div className="w-full">
      <form>
        <div className="flex items-center gap-2">
          <Button
            className="flex items-center justify-between w-fit mr-1.5 py-1 !px-1"
            variant="ghost"
            type="button"
            onClick={handleShowAll}
          >
            <span className="flex items-center gap-1">
              <span className="font-semibold text-xs">Show all</span>
            </span>
          </Button>
          {renderFilters('', '')}
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
          <Funnel size="14" /> Filter ({selectedTypes.length})
        </>
      }
      isDownIcon
    >
      <div className="min-w-[200px]">
        <form>
          <div className="py-2 flex flex-col gap-1">
            {renderFilters('', 'w-[40px]')}
          </div>
          <hr />
          <Button
            className="w-full !rounded-t-none !text-base flex items-center gap-2 pl-4"
            type="button"
            onClick={handleShowAll}
            variant="ghost"
          >
            Reset
          </Button>
        </form>
      </div>
    </Dropdown>
  )
}
