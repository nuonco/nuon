import React from 'react'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export const ConfigSection = ({
  actions,
  children,
  title,
}: {
  actions?: React.ReactNode
  children: React.ReactNode
  title: React.ReactNode
}) => (
  <section className="flex flex-col gap-3">
    <div className="flex items-center justify-between gap-4 min-h-8">
      <Text variant="body" weight="strong">
        {title}
      </Text>
      {actions ? (
        <div className="flex items-center gap-2">{actions}</div>
      ) : null}
    </div>
    {children}
  </section>
)

export type TConfigProperty = {
  label: string
  value?: React.ReactNode
}

export const ConfigProperties = ({
  properties,
}: {
  properties: TConfigProperty[]
}) => {
  const rows = properties.filter(
    ({ value }) => value !== undefined && value !== null && value !== ''
  )

  if (!rows.length) return null

  return (
    <dl className="grid grid-cols-[minmax(8rem,max-content)_1fr] rounded-md border">
      {rows.map(({ label, value }, index) => {
        const isLast = index === rows.length - 1

        return (
          <React.Fragment key={label}>
            <dt className={cn('px-4 py-3', !isLast && 'border-b')}>
              <Text variant="subtext" theme="neutral">
                {label}
              </Text>
            </dt>
            <dd
              className={cn(
                'px-4 py-3 min-w-0 break-all flex items-center',
                !isLast && 'border-b'
              )}
            >
              {typeof value === 'string' || typeof value === 'number' ? (
                <Text variant="subtext" family="mono" className="break-all">
                  {value}
                </Text>
              ) : (
                value
              )}
            </dd>
          </React.Fragment>
        )
      })}
    </dl>
  )
}
