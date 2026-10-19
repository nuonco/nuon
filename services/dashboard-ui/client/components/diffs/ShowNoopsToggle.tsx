import { Toggle } from '@/components/common/form/Toggle'

export const ShowNoopsToggle = ({
  checked,
  onChange,
}: {
  checked: boolean
  onChange: (checked: boolean) => void
}) => (
  <Toggle checked={checked} onChange={onChange} label="Show no-ops" />
)
