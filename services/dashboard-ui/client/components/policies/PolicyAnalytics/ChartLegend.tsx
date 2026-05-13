import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IChartLegendItem {
  color: string
  label: string
}

export interface IChartLegend {
  items: IChartLegendItem[]
  className?: string
}

/**
 * Inline chart legend rendered outside the SVG so it never shifts the inner
 * plot area as bars/cards resize. Pairs with the design-system `<Text>`
 * subtext variant to stay typographically consistent with the rest of the
 * dashboard.
 */
export const ChartLegend = ({ items, className }: IChartLegend) => (
  <div className={cn('flex items-center gap-x-4 gap-y-1 flex-wrap', className)}>
    {items.map((item) => (
      <div key={item.label} className="flex items-center gap-1.5">
        <span
          aria-hidden
          className="inline-block w-2.5 h-2.5 rounded-[2px] shrink-0"
          style={{ backgroundColor: item.color }}
        />
        <Text variant="subtext" theme="neutral">
          {item.label}
        </Text>
      </div>
    ))}
  </div>
)
