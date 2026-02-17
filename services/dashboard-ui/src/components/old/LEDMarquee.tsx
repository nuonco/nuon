'use client'

import React, { useRef, useEffect } from 'react'

interface LEDMarqueeProps {
  text?: string
  dotColor?: string
  bgColor?: string
}

// 5x7 Dot Matrix Font - each character is a 5-column, 7-row grid
const FONT: Record<string, number[]> = {
  A: [0x7c, 0x12, 0x11, 0x12, 0x7c],
  B: [0x7f, 0x49, 0x49, 0x49, 0x36],
  C: [0x3e, 0x41, 0x41, 0x41, 0x22],
  D: [0x7f, 0x41, 0x41, 0x22, 0x1c],
  E: [0x7f, 0x49, 0x49, 0x49, 0x41],
  F: [0x7f, 0x09, 0x09, 0x09, 0x01],
  G: [0x3e, 0x41, 0x49, 0x49, 0x7a],
  H: [0x7f, 0x08, 0x08, 0x08, 0x7f],
  I: [0x00, 0x41, 0x7f, 0x41, 0x00],
  J: [0x20, 0x40, 0x41, 0x3f, 0x01],
  K: [0x7f, 0x08, 0x14, 0x22, 0x41],
  L: [0x7f, 0x40, 0x40, 0x40, 0x40],
  M: [0x7f, 0x02, 0x0c, 0x02, 0x7f],
  N: [0x7f, 0x04, 0x08, 0x10, 0x7f],
  O: [0x3e, 0x41, 0x41, 0x41, 0x3e],
  P: [0x7f, 0x09, 0x09, 0x09, 0x06],
  Q: [0x3e, 0x41, 0x51, 0x21, 0x5e],
  R: [0x7f, 0x09, 0x19, 0x29, 0x46],
  S: [0x46, 0x49, 0x49, 0x49, 0x31],
  T: [0x01, 0x01, 0x7f, 0x01, 0x01],
  U: [0x3f, 0x40, 0x40, 0x40, 0x3f],
  V: [0x1f, 0x20, 0x40, 0x20, 0x1f],
  W: [0x3f, 0x40, 0x38, 0x40, 0x3f],
  X: [0x63, 0x14, 0x08, 0x14, 0x63],
  Y: [0x07, 0x08, 0x70, 0x08, 0x07],
  Z: [0x61, 0x51, 0x49, 0x45, 0x43],
  '0': [0x3e, 0x51, 0x49, 0x45, 0x3e],
  '1': [0x00, 0x42, 0x7f, 0x40, 0x00],
  '2': [0x42, 0x61, 0x51, 0x49, 0x46],
  '3': [0x21, 0x41, 0x45, 0x4b, 0x31],
  '4': [0x18, 0x14, 0x12, 0x7f, 0x10],
  '5': [0x27, 0x45, 0x45, 0x45, 0x39],
  '6': [0x3c, 0x4a, 0x49, 0x49, 0x30],
  '7': [0x01, 0x71, 0x09, 0x05, 0x03],
  '8': [0x36, 0x49, 0x49, 0x49, 0x36],
  '9': [0x06, 0x49, 0x49, 0x29, 0x1e],
  ' ': [0x00, 0x00, 0x00, 0x00, 0x00],
  ':': [0x00, 0x36, 0x36, 0x00, 0x00],
  '-': [0x08, 0x08, 0x08, 0x08, 0x08],
  '.': [0x00, 0x60, 0x60, 0x00, 0x00],
  ',': [0x00, 0x48, 0x30, 0x00, 0x00],
  "'": [0x00, 0x03, 0x03, 0x00, 0x00],
}

export const LEDMarquee: React.FC<LEDMarqueeProps> = ({
  text = 'See What Your Customers See',
  dotColor = '#22d3ee',
  bgColor = '#0a0a0a',
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    const wrapper = wrapperRef.current
    if (!canvas || !wrapper) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = window.devicePixelRatio || 1

    const DOT_SIZE = 2.5
    const DOT_GAP = 1
    const CHAR_GAP = 2
    const CELL_SIZE = DOT_SIZE + DOT_GAP
    const CHAR_WIDTH = 5
    const CHAR_HEIGHT = 7
    const PADDING_Y = 4
    const PADDING_ROWS = 2
    const TOTAL_ROWS = CHAR_HEIGHT + PADDING_ROWS * 2

    function getTextWidth(str: string) {
      let width = 0
      for (let i = 0; i < str.length; i++) {
        width += CHAR_WIDTH * CELL_SIZE + CHAR_GAP
      }
      return width - CHAR_GAP
    }

    // Generate stable inactive dot pattern
    let inactiveDotsPattern: number[][] = []

    function generateInactivePattern(totalCols: number) {
      inactiveDotsPattern = []
      for (let col = 0; col < totalCols; col++) {
        inactiveDotsPattern[col] = []
        for (let row = 0; row < TOTAL_ROWS; row++) {
          inactiveDotsPattern[col][row] = 0.04 + Math.random() * 0.08
        }
      }
    }

    function drawFrame(charCount: number) {
      const totalWidth = wrapper!.clientWidth
      const totalHeight = TOTAL_ROWS * CELL_SIZE + PADDING_Y * 2

      canvas!.width = totalWidth * dpr
      canvas!.height = totalHeight * dpr
      canvas!.style.width = totalWidth + 'px'
      canvas!.style.height = totalHeight + 'px'
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0)

      ctx!.fillStyle = bgColor
      ctx!.fillRect(0, 0, totalWidth, totalHeight)

      const totalCols = Math.ceil(totalWidth / CELL_SIZE)

      if (inactiveDotsPattern.length !== totalCols) {
        generateInactivePattern(totalCols)
      }

      // Draw inactive dots
      for (let col = 0; col < totalCols; col++) {
        for (let row = 0; row < TOTAL_ROWS; row++) {
          const opacity = inactiveDotsPattern[col]?.[row] || 0.06
          ctx!.fillStyle = `rgba(34, 211, 238, ${opacity})`
          const x = col * CELL_SIZE
          const y = PADDING_Y + row * CELL_SIZE
          ctx!.beginPath()
          ctx!.arc(x + DOT_SIZE / 2, y + DOT_SIZE / 2, DOT_SIZE / 2, 0, Math.PI * 2)
          ctx!.fill()
        }
      }

      // Draw active text dots
      const textWidth = getTextWidth(text)
      const startX = (totalWidth - textWidth) / 2
      const textStartRow = PADDING_ROWS

      ctx!.fillStyle = dotColor
      ctx!.shadowColor = dotColor
      ctx!.shadowBlur = 4

      let xOffset = startX
      const charsToShow = Math.min(charCount, text.length)

      for (let i = 0; i < charsToShow; i++) {
        const char = text[i].toUpperCase()
        const charData = FONT[char] || FONT[' ']

        for (let col = 0; col < CHAR_WIDTH; col++) {
          const colData = charData[col] || 0
          for (let row = 0; row < CHAR_HEIGHT; row++) {
            if (colData & (1 << row)) {
              const x = xOffset + col * CELL_SIZE
              const y = PADDING_Y + (textStartRow + row) * CELL_SIZE
              ctx!.beginPath()
              ctx!.arc(
                x + DOT_SIZE / 2,
                y + DOT_SIZE / 2,
                DOT_SIZE / 2,
                0,
                Math.PI * 2,
              )
              ctx!.fill()
            }
          }
        }
        xOffset += CHAR_WIDTH * CELL_SIZE + CHAR_GAP
      }
    }

    // Typing animation state
    let currentChar = 0
    const typingSpeed = 40
    let lastTime = 0
    let animationComplete = false
    let frameId: number | null = null

    function animate(timestamp: number) {
      if (!lastTime) lastTime = timestamp

      const elapsed = timestamp - lastTime

      if (!animationComplete) {
        if (elapsed > typingSpeed) {
          currentChar++
          lastTime = timestamp

          if (currentChar >= text.length) {
            animationComplete = true
          }
        }
        drawFrame(currentChar)
      }

      if (!animationComplete) {
        frameId = requestAnimationFrame(animate)
      }
    }

    // Draw empty state first
    drawFrame(0)

    // Start typing animation after a short delay
    const timer = setTimeout(() => {
      frameId = requestAnimationFrame(animate)
    }, 400)

    // Handle resize
    const resizeObserver = new ResizeObserver(() => {
      inactiveDotsPattern = []
      drawFrame(animationComplete ? text.length : currentChar)
    })
    resizeObserver.observe(wrapper)

    return () => {
      clearTimeout(timer)
      if (frameId) cancelAnimationFrame(frameId)
      resizeObserver.disconnect()
    }
  }, [text, dotColor, bgColor])

  return (
    <div
      ref={wrapperRef}
      style={{
        background: bgColor,
        borderTop: `1px solid rgba(34, 211, 238, 0.15)`,
        borderBottom: `1px solid rgba(34, 211, 238, 0.15)`,
        overflow: 'hidden',
        width: '100%',
      }}
    >
      <canvas
        ref={canvasRef}
        style={{ display: 'block', imageRendering: 'pixelated' as const }}
      />
    </div>
  )
}
