import { useEffect } from 'react'
import { usePageTitle } from '@/hooks/use-page-title'

export const composeTitle = (segments: (string | undefined | null | false)[]): string =>
  segments
    .filter((s): s is string => typeof s === 'string' && s.length > 0)
    .join(' | ')

type IPageTitle =
  | { title: string; segments?: never }
  | { segments: (string | undefined | null | false)[]; title?: never }

export const PageTitle = ({ title, segments }: IPageTitle) => {
  const { updateTitle } = usePageTitle()
  const resolved = segments ? composeTitle(segments) : (title ?? '')
  useEffect(() => {
    updateTitle(resolved)
  }, [resolved])
  return <></>
}
