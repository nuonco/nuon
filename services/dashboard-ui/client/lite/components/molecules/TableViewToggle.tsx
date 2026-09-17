import type { TTableView } from '../../hooks/use-table-view'
import { Icon } from '../atoms/Icon'
import { ToggleButton } from '../atoms/ToggleButton'

export interface ITableViewToggle {
  value: TTableView
  onValueChange: (value: TTableView) => void
  label?: string
  className?: string
}

export const TableViewToggle = ({
  value,
  onValueChange,
  label = 'Collection view',
  className,
}: ITableViewToggle) => (
  <ToggleButton
    value={value}
    onValueChange={onValueChange}
    label={label}
    className={className}
    options={[
      {
        value: 'table',
        label: <Icon variant="TableIcon" size={14} aria-hidden />,
        ariaLabel: 'Table view',
        tooltip: 'Table view',
        iconOnly: true,
      },
      {
        value: 'cards',
        label: <Icon variant="SquaresFourIcon" size={14} aria-hidden />,
        ariaLabel: 'Card view',
        tooltip: 'Card view',
        iconOnly: true,
      },
    ]}
  />
)
