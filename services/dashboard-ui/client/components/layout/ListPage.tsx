import type { ReactNode } from 'react'
import { PageContent } from './PageContent'
import { PageLayout } from './PageLayout'
import { PageSection } from './PageSection'
import { SectionHeader, type ISectionHeader } from './SectionHeader'

export interface IListPage extends ISectionHeader {
  children: ReactNode
  createAction?: ReactNode
}

export const ListPage = ({
  actions,
  children,
  className,
  createAction,
  description,
  status,
  title,
  variant = 'section',
}: IListPage) => {
  const headerActions =
    actions || createAction ? (
      <>
        {actions}
        {createAction}
      </>
    ) : undefined

  const header = (
    <SectionHeader
      actions={headerActions}
      description={description}
      status={status}
      title={title}
      variant={variant}
    />
  )

  if (variant === 'page') {
    return (
      <PageLayout>
        {header}
        <PageContent>
          <PageSection className={className}>{children}</PageSection>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <PageSection className={className}>
      {header}
      {children}
    </PageSection>
  )
}
