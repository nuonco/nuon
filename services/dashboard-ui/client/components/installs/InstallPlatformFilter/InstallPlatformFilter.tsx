import React from 'react'
import { useSearchParams } from 'react-router'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import type { TCloudPlatform } from '@/types'

const FILTER_OPTIONS: TCloudPlatform[] = ['aws', 'azure', 'gcp', 'unknown']
const PLATFORM_LABEL: Record<TCloudPlatform, string> = {
  aws: 'AWS',
  azure: 'Azure',
  gcp: 'GCP',
  unknown: 'Unknown',
}
const PARAM = 'cloud_platform'

export const InstallPlatformFilter = () => {
  const [searchParams, setSearchParams] = useSearchParams()

  const param = searchParams.get(PARAM)
  const allSelected = !param
  const selected: TCloudPlatform[] = allSelected
    ? FILTER_OPTIONS
    : param
        .split(',')
        .map((v) => v.trim().toLowerCase())
        .filter((v): v is TCloudPlatform =>
          FILTER_OPTIONS.includes(v as TCloudPlatform)
        )

  const setPlatformsInUrl = (values: TCloudPlatform[]) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        if (values.length === 0 || values.length === FILTER_OPTIONS.length) {
          params.delete(PARAM)
        } else {
          params.set(PARAM, values.join(','))
        }
        params.delete('offset')
        return params
      },
      { replace: true }
    )
  }

  const handleToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value as TCloudPlatform
    if (e.target.checked) {
      setPlatformsInUrl(Array.from(new Set([...selected, value])))
    } else {
      setPlatformsInUrl(selected.filter((v) => v !== value))
    }
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    setPlatformsInUrl([e.currentTarget.value as TCloudPlatform])
  }

  const handleShowAll = () => setPlatformsInUrl(FILTER_OPTIONS)

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id="platform-filter"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" />
          Platform{!allSelected ? ` (${selected.length})` : ''}
        </>
      }
    >
      <Menu className="min-w-64 max-h-80 overflow-y-auto">
        <Text variant="label" theme="neutral" className="px-1">
          Filter by platform
        </Text>

        {FILTER_OPTIONS.map((value) => {
          const isOnlySelected = selected.length === 1 && selected[0] === value
          return (
            <div className="flex items-center space-x-2" key={value}>
              <CheckboxInputWithButton
                buttonProps={{
                  className:
                    '!p-1 flex items-center justify-between group w-full',
                  children: (
                    <>
                      <span className="flex items-center gap-2">
                        <CloudPlatform
                          platform={value}
                          variant="subtext"
                          colorVariant="mono"
                          displayVariant="icon-only"
                          iconSize="16"
                        />
                        <span className="font-semibold text-xs">
                          {PLATFORM_LABEL[value]}
                        </span>
                      </span>
                      <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                        {isOnlySelected ? 'Reset' : 'Only'}
                      </span>
                    </>
                  ),
                  type: 'button',
                  variant: 'ghost',
                  value,
                  onClick: isOnlySelected ? handleShowAll : handleOnly,
                }}
                className="w-full"
                name={value}
                onChange={handleToggle}
                checked={selected.includes(value)}
                value={value}
              />
            </div>
          )
        })}

        <hr />

        <Button
          className="w-full !p-1 shrink-0"
          type="button"
          onClick={handleShowAll}
          size="sm"
          variant="ghost"
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
