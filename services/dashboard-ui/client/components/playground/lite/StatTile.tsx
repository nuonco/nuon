import { Block } from './Block'
import { labelWidth } from './utils'

export interface IStatTile {
  label: string
  valueWidth?: number
}

export const StatTile = ({ label, valueWidth = 64 }: IStatTile) => (
  <div className="flex flex-col gap-3 rounded-lg bg-cool-grey-100 dark:bg-dark-grey-800 p-4">
    <Block
      className="h-[10px] opacity-60"
      style={{ width: labelWidth(label) }}
      title={label}
      text={label}
    />
    <Block className="h-[28px]" style={{ width: valueWidth }} title={label} />
    <Block className="w-[80px] h-[8px] opacity-40" title="trend" />
  </div>
)
