import type { ReactNode } from 'react'
import { PageContent } from './PageContent'
import { PageLayout } from './PageLayout'
import { PageSection } from './PageSection'
import { SectionHeader, type ISectionHeader } from './SectionHeader'

export interface IListPage extends ISectionHeader {
  children: ReactNode
  createAction?: ReactNode
  filters?: ReactNode
}

export const ListPage = ({
  actions,
  children,
  className,
  createAction,
  description,
  filters,
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

  const body = (
    <>
      {filters ? (
        <div className="flex flex-wrap items-center gap-3">{filters}</div>
      ) : null}
      {children}
    </>
  )

  if (variant === 'page') {
    return (
      <PageLayout>
        <SectionHeader
          actions={headerActions}
          description={description}
          status={status}
          title={title}
          variant="page"
        />
        <PageContent>
          <PageSection className={className}>{body}</PageSection>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <PageSection className={className}>
      <SectionHeader
        actions={headerActions}
        description={description}
        status={status}
        title={title}
        variant="section"
      />
      {body}
    </PageSection>
  )
}
