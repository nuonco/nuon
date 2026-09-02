import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

interface IShowPreviewRuns {
  showPreviews: boolean
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

export const ShowPreviewRuns = ({
  showPreviews,
  onChange,
}: IShowPreviewRuns) => {
  return (
    <CheckboxInput
      labelProps={{ labelText: 'Preview runs' }}
      checked={showPreviews}
      onChange={onChange}
    />
  )
}
