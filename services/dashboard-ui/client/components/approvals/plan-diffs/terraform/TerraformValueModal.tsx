import { Icon } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'
import { Modal } from '@/components/surfaces/Modal'
import { TerraformValueCodeBlock } from './TerraformValueCodeBlock'

export const TerraformValueModal = ({
  isBefore = false,
  valueKey,
  value,
}: {
  isBefore?: boolean
  valueKey: string
  value: string
}) => {
  return (
    <Modal
      size="half"
      heading={`${isBefore ? 'Before:' : 'After:'} ${valueKey}`}
      triggerButton={{
        children: (
          <Tooltip
            position={isBefore ? 'right' : 'left'}
            tipContent="View details"
          >
            <Icon variant="MagnifyingGlassPlus" size="14" />
          </Tooltip>
        ),
        className: '!p-0.5 !leading-none !h-auto',
        size: 'sm',
      }}
    >
      <TerraformValueCodeBlock value={value} />
    </Modal>
  )
}
