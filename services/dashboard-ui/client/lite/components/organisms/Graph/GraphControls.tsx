import { useReactFlow } from '@xyflow/react'
import { useMediaQuery } from '../../../hooks/use-media-query'
import { Button } from '../../atoms/Button'
import { Icon } from '../../atoms/Icon'

export const GraphControls = () => {
  const { zoomIn, zoomOut, fitView } = useReactFlow()
  const reducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const fitDuration = reducedMotion ? 0 : 200

  return (
    <div
      className="absolute right-3 bottom-3 z-10 flex items-center gap-1"
      role="toolbar"
      aria-label="Graph controls"
    >
      <Button
        variant="ghost"
        size="sm"
        iconOnly
        aria-label="Zoom in"
        onClick={() => zoomIn({ duration: fitDuration })}
      >
        <Icon variant="PlusIcon" size={14} />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        iconOnly
        aria-label="Zoom out"
        onClick={() => zoomOut({ duration: fitDuration })}
      >
        <Icon variant="MinusIcon" size={14} />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        iconOnly
        aria-label="Fit to view"
        onClick={() => fitView({ padding: 0.2, duration: fitDuration })}
      >
        <Icon variant="CornersOutIcon" size={14} />
      </Button>
    </div>
  )
}
