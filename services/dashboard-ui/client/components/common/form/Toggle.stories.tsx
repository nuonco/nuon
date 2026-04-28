import { useState } from 'react'
import { Toggle } from './Toggle'

export default {
  title: 'Common/Form/Toggle',
}

export const Default = () => {
  const [checked, setChecked] = useState(false)
  return <Toggle checked={checked} onChange={setChecked} />
}

export const WithLabel = () => {
  const [checked, setChecked] = useState(false)
  return <Toggle checked={checked} onChange={setChecked} label="Auto approval" />
}

export const Checked = () => {
  const [checked, setChecked] = useState(true)
  return <Toggle checked={checked} onChange={setChecked} label="Auto approval" />
}

export const WithDescription = () => {
  const [checked, setChecked] = useState(false)
  return (
    <Toggle
      checked={checked}
      onChange={setChecked}
      label="Auto approval"
      description="Automatically approve all workflows"
    />
  )
}

export const Disabled = () => (
  <div className="flex flex-col gap-4">
    <Toggle checked={false} onChange={() => {}} label="Disabled off" disabled />
    <Toggle checked={true} onChange={() => {}} label="Disabled on" disabled />
  </div>
)
