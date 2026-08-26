import type { ReactNode } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { cn } from '@/utils/classnames'
import { SectionHeaderRow, type TSectionHeaderVariant } from './SectionHeader'

export interface IDetailHeader {
  actions?: ReactNode
  backLink?: boolean
  children?: ReactNode
  className?: string
  description?: ReactNode
  icon?: ReactNode
  id?: ReactNode
  identity?: ReactNode
  loading?: boolean
  loadingWidth?: number
  metadata?: ReactNode
  status?: ReactNode
  title: ReactNode
  variant?: TSectionHeaderVariant
}

export const DetailHeader = ({
  actions,
  backLink = true,
  children,
  className,
  description,
  icon,
  id,
  identity,
  loading,
  loadingWidth,
  metadata,
  status,
  title,
  variant = 'section',
}: IDetailHeader) => {
  const titleContent = icon ? (
    <span className="flex items-center gap-2 min-w-0">
      {icon}
      {title}
    </span>
  ) : (
    title
  )

  const content = (
    <>
      <div className="flex flex-col gap-2 w-full">
        {backLink ? <BackLink className="mb-2" /> : null}
        <SectionHeaderRow
          actions={actions}
          description={description}
          loading={loading}
          loadingWidth={loadingWidth}
          status={status}
          title={titleContent}
          variant={variant}
        />
        {id || identity ? (
          <div className="flex flex-wrap items-center gap-4 min-w-0">
            {id ? <ID loading={loading}>{id}</ID> : null}
            {identity}
          </div>
        ) : null}
      </div>

      {metadata ? (
        <Card>
          <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
            {metadata}
          </div>
        </Card>
      ) : null}

      {children}
    </>
  )

  return (
    <header
      className={cn(
        'flex flex-col gap-6 w-full',
        variant === 'page' && 'shrink-0 p-4 md:p-6',
        className
      )}
    >
      {content}
    </header>
  )
}
