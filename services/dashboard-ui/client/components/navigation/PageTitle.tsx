import { useEffect } from 'react'
import { usePageTitle } from '@/hooks/use-page-title'

export const PageTitle = ({ title }: { title: string }) => {
  const { updateTitle } = usePageTitle()
  useEffect(() => {
    updateTitle(title)
  }, [title])
  return <></>
}
