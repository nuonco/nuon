'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { type ChangeEvent } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'

export const ShowDriftScan = () => {
  const router = useRouter()
  const searchParams = useSearchParams()

  const showDrifts = searchParams.get('drifts') !== 'false'

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked

    const params = new URLSearchParams(searchParams.toString())
    params.set('drifts', checked ? 'true' : 'false')
    router.replace(`?${params.toString()}`)
  }

  return (
    <>
      <CheckboxInput
        labelProps={{ labelText: 'Drift scan' }}
        checked={showDrifts}
        onChange={handleChange}
      />
    </>
  )
}
