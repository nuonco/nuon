'use client'

import { useEffect, useState, useRef } from 'react'

type CursorState = 'default' | 'text' | 'project' | 'terminal'

export const CustomCursor = () => {
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  const [cursorPos, setCursorPos] = useState({ x: 0, y: 0 })
  const [cursorState, setCursorState] = useState<CursorState>('default')
  const [isVisible, setIsVisible] = useState(false)
  const [scale, setScale] = useState(1)
  const requestRef = useRef<number | undefined>(undefined)

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      setMousePos({ x: e.clientX, y: e.clientY })
      if (!isVisible) setIsVisible(true)
    }

    const handleMouseEnter = () => setIsVisible(true)
    const handleMouseLeave = () => setIsVisible(false)

    // Check cursor state based on element type
    const checkCursorState = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      
      // Check for terminal hover (data-cursor-terminal)
      if (target.closest('[data-cursor-terminal]')) {
        setCursorState('terminal')
        return
      }
      
      // Check for project hover (data-cursor-project)
      if (target.closest('[data-cursor-project]')) {
        setCursorState('project')
        return
      }
      
      // Check for text/link hover
      if (target.closest('a') || target.closest('button') || target.closest('[data-cursor-hover]')) {
        setCursorState('text')
        return
      }
      
      setCursorState('default')
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mousemove', checkCursorState)
    document.addEventListener('mouseenter', handleMouseEnter)
    document.addEventListener('mouseleave', handleMouseLeave)

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mousemove', checkCursorState)
      document.removeEventListener('mouseenter', handleMouseEnter)
      document.removeEventListener('mouseleave', handleMouseLeave)
    }
  }, [isVisible])

  // Smooth cursor following with lerp
  useEffect(() => {
    const animate = () => {
      setCursorPos((prev) => {
        const dx = mousePos.x - prev.x
        const dy = mousePos.y - prev.y
        return {
          x: prev.x + dx * 0.15, // Smoother lerp factor
          y: prev.y + dy * 0.15,
        }
      })

      // Smooth scale transition based on cursor state
      setScale((prev) => {
        let target = 1
        if (cursorState === 'text') target = 1.3
        if (cursorState === 'project') target = 1.6
        if (cursorState === 'terminal') target = 1.2
        return prev + (target - prev) * 0.15
      })

      requestRef.current = requestAnimationFrame(animate)
    }

    requestRef.current = requestAnimationFrame(animate)
    return () => {
      if (requestRef.current) {
        cancelAnimationFrame(requestRef.current)
      }
    }
  }, [mousePos, cursorState])

  const size = cursorState === 'default' ? 10 : 28 * scale
  const isHovering = cursorState !== 'default'

  return (
    <>
      {/* Hide default cursor */}
      <style jsx global>{`
        *,
        *::before,
        *::after,
        html,
        body,
        a,
        button,
        input,
        textarea {
          cursor: none !important;
        }
      `}</style>

      {/* Custom cursor */}
      <div
        className={`fixed pointer-events-none z-[9999] transition-opacity duration-300 ${
          isVisible ? 'opacity-100' : 'opacity-0'
        }`}
        style={{
          left: cursorPos.x,
          top: cursorPos.y,
          transform: 'translate(-50%, -50%)',
          willChange: 'transform',
        }}
      >
        {/* Default state: simple solid circle */}
        {cursorState === 'default' && (
          <div
            className="rounded-full"
            style={{
              width: size,
              height: size,
              backgroundColor: 'rgba(255, 255, 255, 0.9)',
              transition: 'all 0.2s ease',
            }}
          />
        )}

        {/* Terminal state: CLI-style block cursor */}
        {cursorState === 'terminal' && (
          <div
            className="animate-pulse"
            style={{
              width: '12px',
              height: '20px',
              backgroundColor: 'rgba(255, 255, 255, 0.9)',
              transition: 'all 0.2s ease',
            }}
          />
        )}

        {/* Hover states: clean circle with subtle border */}
        {(cursorState === 'text' || cursorState === 'project') && (
          <div
            className="rounded-full relative"
            style={{
              width: size,
              height: size,
              border: `1.5px solid rgba(255, 255, 255, 0.8)`,
              backgroundColor: 'rgba(255, 255, 255, 0.1)',
              transition: 'all 0.3s ease',
            }}
          >
            {/* Inner dot for all hover states */}
            <div
              className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full"
              style={{
                width: cursorState === 'project' ? '8px' : '6px',
                height: cursorState === 'project' ? '8px' : '6px',
                backgroundColor: 'rgba(255, 255, 255, 0.9)',
                transition: 'all 0.3s ease',
              }}
            />
          </div>
        )}
      </div>
    </>
  )
}

