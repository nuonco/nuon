export default {
  title: 'Datadog/SiteInput',
}

import { useState } from 'react'
import { SiteInput } from './SiteInput'

export const Default = () => {
  const [value, setValue] = useState('us1')
  return (
    <div className="w-[420px] p-4">
      <SiteInput value={value} onChange={setValue} />
      <div className="mt-4 text-xs text-stone-500">Value: {value || '(empty)'}</div>
    </div>
  )
}

export const StartingWithCustomURL = () => {
  const [value, setValue] = useState('https://datadog.internal.example.com')
  return (
    <div className="w-[420px] p-4">
      <SiteInput value={value} onChange={setValue} />
      <div className="mt-4 text-xs text-stone-500">Value: {value || '(empty)'}</div>
    </div>
  )
}

export const Disabled = () => {
  const [value, setValue] = useState('eu1')
  return (
    <div className="w-[420px] p-4">
      <SiteInput value={value} onChange={setValue} disabled />
    </div>
  )
}
