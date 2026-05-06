'use client'

import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface ILogToolFilter {
  title?: string
  filters: {
    tool: TLogFiltersProps['tool']
    setTool: TLogFiltersProps['setTool']
    availableTools: TLogFiltersProps['availableTools']
  }
}

// Single-value filter for log_attributes['nuon.tool']. Shows tools observed
// in the loaded logs plus the currently-selected value (so a tool selected
// via URL state is still visible if the page hasn't loaded matching records
// yet).
export const LogToolFilter = ({ title = 'Tool', filters }: ILogToolFilter) => {
  const { tool, setTool, availableTools } = filters

  const options = Array.from(
    new Set<string>([...availableTools, ...(tool ? [tool] : [])])
  ).sort()

  return (
    <Dropdown
      buttonText={`${title} (${tool || 'all'})`}
      icon={<Icon variant="FunnelIcon" size="14" />}
      iconAlignment="left"
      id={`${title}-filter`}
      alignment="right"
      variant="ghost"
    >
      <Menu className="min-w-56">
        {options.length === 0 && (
          <div className="px-1 py-2 text-xs text-neutral-500">No tools</div>
        )}
        {options.map((option) => (
          <Button
            key={option}
            className="!p-1 flex items-center justify-between group w-full"
            variant="ghost"
            size="sm"
            onClick={() => setTool(tool === option ? '' : option)}
          >
            <span className="flex items-center gap-1.5 h-full">{option}</span>
            {tool === option && (
              <Icon variant="Check" size="12" />
            )}
          </Button>
        ))}

        <hr />
        <Button
          className="w-full !p-1 shrink-0"
          variant="ghost"
          size="sm"
          onClick={() => setTool('')}
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
