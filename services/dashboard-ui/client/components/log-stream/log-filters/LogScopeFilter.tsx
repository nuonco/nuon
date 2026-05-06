'use client'

import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface ILogScopeFilter {
  title?: string
  filters: {
    selectedScopes: TLogFiltersProps['selectedScopes']
    availableScopes: TLogFiltersProps['availableScopes']
    handleScopeInputToggle: TLogFiltersProps['handleScopeInputToggle']
    handleScopeButtonClick: TLogFiltersProps['handleScopeButtonClick']
    handleScopeReset: TLogFiltersProps['handleScopeReset']
  }
}

export const LogScopeFilter = ({ title = 'Scope', filters }: ILogScopeFilter) => {
  const {
    selectedScopes,
    availableScopes,
    handleScopeInputToggle,
    handleScopeButtonClick,
    handleScopeReset,
  } = filters

  const options = Array.from(
    new Set<string>([...availableScopes, ...selectedScopes])
  ).sort()

  const getButtonText = (scope: string) => {
    if (selectedScopes.size === 1 && selectedScopes.has(scope)) {
      return 'Reset'
    }
    return 'Only'
  }

  return (
    <Dropdown
      buttonText={`${title} (${selectedScopes.size === 0 ? 'all' : selectedScopes.size})`}
      icon={<Icon variant="FunnelIcon" size="14" />}
      iconAlignment="left"
      id={`${title}-filter`}
      alignment="right"
      variant="ghost"
    >
      <Menu className="min-w-56">
        {options.length === 0 && (
          <div className="px-1 py-2 text-xs text-neutral-500">No scopes</div>
        )}
        {options.map((scope) => (
          <div key={scope} className="flex items-center space-x-2">
            <input
              type="checkbox"
              id={`${title.toLowerCase().replace(' ', '-')}-scope-${scope}`}
              checked={selectedScopes.has(scope)}
              onChange={() => handleScopeInputToggle(scope)}
              className="focus:ring-primary-500 focus:border-primary-500 accent-primary-600"
            />
            <Button
              className="!p-1 flex items-center justify-between group w-full"
              variant="ghost"
              size="sm"
              onClick={() => handleScopeButtonClick(scope)}
            >
              <span className="flex items-center gap-1.5 h-full">{scope}</span>
              <span className="hidden group-hover:inline-flex text-xs">
                {getButtonText(scope)}
              </span>
            </Button>
          </div>
        ))}

        <hr />
        <Button
          className="w-full !p-1 shrink-0"
          variant="ghost"
          size="sm"
          onClick={handleScopeReset}
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
