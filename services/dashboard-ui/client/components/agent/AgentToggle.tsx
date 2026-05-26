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
        className="group !rounded-full !h-12 !w-12 hover:!w-40 justify-center shadow-lg !transition-all duration-fast ease-cubic overflow-hidden !gap-0"
      >
        <Icon variant="SparkleIcon" size={24} />
        <span className="opacity-0 w-0 whitespace-nowrap transition-all duration-fast ease-cubic group-hover:opacity-100 group-hover:w-full group-hover:ml-3">
          Nuon agent
        </span>
      </Button>
    </div>
  )
}
