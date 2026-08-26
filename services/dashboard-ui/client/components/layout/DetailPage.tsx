import type { ReactNode } from 'react'
import { TabNav, type ITabNav } from '@/components/navigation/TabNav'
import { cn } from '@/utils/classnames'
import { PageContent } from './PageContent'
import { PageLayout } from './PageLayout'
import { PageSection } from './PageSection'

export type TDetailPageVariant = 'page' | 'section'

export interface IDetailPage {
  banners?: ReactNode
  children: ReactNode
  className?: string
  header: ReactNode
  tabNav?: ITabNav
  variant?: TDetailPageVariant
}

export const DetailPage = ({
  banners,
  children,
  className,
  header,
  tabNav,
  variant = 'section',
}: IDetailPage) => {
  const body = (
    <PageSection className={cn('@container', className)}>
      {header}
      {banners}
      {tabNav ? <TabNav {...tabNav} /> : null}
      {children}
    </PageSection>
  )

  if (variant === 'page') {
    return (
      <PageLayout>
        <PageContent>{body}</PageContent>
      </PageLayout>
    )
  }

  return body
}
