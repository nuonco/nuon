import React from 'react'
import { cn } from '@/utils/classnames'
import { EmptyState, type IEmptyState } from './EmptyState/EmptyState'
import { Skeleton } from './Skeleton'
import { Text } from './Text'

export interface IPropertyGridColumn<T> {
  key: keyof T
  header: string
  render?: (value: T[keyof T], item: T, key: keyof T) => React.ReactNode
  className?: string
}

export interface IPropertyGrid<T = Record<string, any>> extends React.HTMLAttributes<HTMLDivElement> {
  values: T[]
  columns?: IPropertyGridColumn<T>[]
  emptyStateProps?: IEmptyState
  gridTemplate?: string  // Custom grid-template-columns CSS value
}

export const PropertyGrid = <T extends Record<string, any>>({
  className,
  values,
  columns,
  gridTemplate,
  emptyStateProps = { variant: 'table', size: 'sm' },
  ...props
}: IPropertyGrid<T>) => {
  // Auto-detect columns from first object if not provided
  const detectedColumns: IPropertyGridColumn<T>[] = React.useMemo(() => {
    if (columns) return columns
    
    if (!values?.length) return []
    
    const firstItem = values[0]
    return Object.keys(firstItem).map(key => ({
      key: key as keyof T,
      header: key.charAt(0).toUpperCase() + key.slice(1).replace(/([A-Z])/g, ' $1').trim()
    }))
  }, [values, columns])

  const gridColumns = detectedColumns.length
  
  // Use custom gridTemplate or create smart defaults
  const gridColsClass = gridTemplate || (
    gridColumns === 1
      ? '1fr'
      : gridColumns === 2 
      ? 'max-content 1fr'  // Two columns: first fits content, second expands (good for key-value pairs)
      : `repeat(${gridColumns}, minmax(120px, 1fr))`  // Multiple columns: all flexible with 120px minimum
  )

  if (!values?.length) {
    return <EmptyState {...emptyStateProps} />
  }

  return (
    <div
      className={cn('grid gap-0', className)}
      style={{
        gridTemplateColumns: gridColsClass
      }}
      {...props}
    >
      {/* Header row */}
      {detectedColumns.map((column, index) => (
        <Text
          key={String(column.key)}
          className={cn(
            'py-2 border-b font-medium',
            index > 0 && 'pl-8',
            column.className
          )}
          variant="subtext"
          theme="neutral"
        >
          {column.header}
        </Text>
      ))}

      {/* Data rows */}
      {values.map((item, itemIndex) => {
        const isLast = itemIndex === values.length - 1

        return detectedColumns.map((column, columnIndex) => {
          const value = item[column.key]
          const renderedValue = column.render 
            ? column.render(value, item, column.key)
            : value

          return (
            <div
              key={`${itemIndex}-${String(column.key)}`}
              className={cn(
                'py-2 break-all flex items-center',
                columnIndex > 0 && 'pl-8',
                !isLast && 'border-b',
                column.className
              )}
            >
              {React.isValidElement(renderedValue) ? (
                renderedValue
              ) : (
                <Text
                  variant="subtext"
                  family="mono"
                  className="break-all !w-full"
                >
                  {renderedValue !== null && renderedValue !== undefined ? (
                    String(renderedValue)
                  ) : (
                    <Text variant="subtext" theme="neutral">
                      —
                    </Text>
                  )}
                </Text>
              )}
            </div>
          )
        })
      })}
    </div>
  )
}

export const PropertyGridSkeleton = ({ 
  count = 5, 
  columns = 2 
}: { 
  count?: number
  columns?: number 
}) => {
  // Use same flexible grid logic as main component
  const gridColsClass = columns === 1
    ? '1fr'
    : columns === 2 
    ? 'max-content 1fr'  // Two columns: first fits content, second expands
    : `repeat(${columns}, minmax(120px, 1fr))`  // Multiple columns: all flexible with 120px minimum

  return (
    <div
      className="grid gap-0"
      style={{
        gridTemplateColumns: gridColsClass
      }}
    >
      {/* Header skeleton */}
      {Array.from({ length: columns }).map((_, columnIndex) => (
        <div 
          key={`header-${columnIndex}`}
          className={cn('py-2 border-b', columnIndex > 0 && 'pl-8')}
        >
          <Skeleton height="17px" width="80px" />
        </div>
      ))}

      {/* Data skeleton rows */}
      {Array.from({ length: count }).map((_, rowIndex) => {
        const isLast = rowIndex === count - 1

        return Array.from({ length: columns }).map((_, columnIndex) => (
          <div
            key={`${rowIndex}-${columnIndex}`}
            className={cn(
              'py-2',
              columnIndex > 0 && 'pl-8',
              !isLast && 'border-b'
            )}
          >
            <Skeleton height="17px" width={columnIndex === 0 ? '120px' : '60%'} />
          </div>
        ))
      })}
    </div>
  )
}