import { useEffect } from 'react'
import { usePageTitle } from '@/hooks/use-page-title'

type Segment = string | undefined | null | false

export const composeTitle = (segments: Segment[]): string =>
  segments
    .filter((s): s is string => typeof s === 'string' && s.length > 0)
    .join(' | ')

type IPageTitle =
  | { title: string; segments?: never }
  | { segments: Segment[]; title?: never }

export const PageTitle = ({ title, segments }: IPageTitle) => {
  const { updateTitle } = usePageTitle()
  const resolved = segments ? composeTitle(segments) : (title ?? '')
  useEffect(() => {
    updateTitle(resolved)
  }, [resolved])
  return null
}
