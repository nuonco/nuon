import { Icon } from './Icon'
import { ToggleButton } from './ToggleButton'

export const DEPENDENCY_VIEW_STORAGE_KEY = 'nuon:dependency-graph-view'
export const DEPENDENCY_VIEW_MODES = ['graph', 'table'] as const
export type TDependencyViewMode = (typeof DEPENDENCY_VIEW_MODES)[number]

export interface IDependencyViewToggle {
  value: TDependencyViewMode
  onChange: (value: TDependencyViewMode) => void
}

export const DependencyViewToggle = ({
  value,
  onChange,
}: IDependencyViewToggle) => (
  <ToggleButton
    value={value}
    onChange={onChange}
    options={[
      {
        value: 'graph',
        label: <Icon variant="GraphIcon" size={16} />,
        ariaLabel: 'Graph view',
      },
      {
        value: 'table',
        label: <Icon variant="TableIcon" size={16} />,
        ariaLabel: 'Table view',
      },
    ]}
  />
)
