import { useAgent } from '@/hooks/use-agent'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { AgentPanel } from './AgentPanel'

export function AgentToggle() {
  const { isOpen, setIsOpen } = useAgent()
  const { addPanel } = useSurfaces()

  const handleOpen = () => {
    setIsOpen(true)
    addPanel(<AgentPanel onClose={() => setIsOpen(false)} />)
  }

  if (isOpen) return null

  return (
    <div className="fixed bottom-5 right-5 z-40">
      <Button
        variant="primary"
        onClick={handleOpen}
        title="Open agent"
        className="!rounded-full !h-12 !w-12 !p-0 justify-center shadow-lg"
      >
        <Icon variant="SparkleIcon" size={20} />
      </Button>
    </div>
  )
}
