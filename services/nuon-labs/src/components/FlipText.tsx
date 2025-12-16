'use client'

import { useState, useEffect, useCallback } from 'react'

interface FlipTextProps {
  text: string
  isVisible: boolean
}

const CHARACTERS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 '

export const FlipText = ({ text, isVisible }: FlipTextProps) => {
  const [displayText, setDisplayText] = useState(text.split('').map(() => ' '))
  const [isAnimating, setIsAnimating] = useState(false)
  const [isHovering, setIsHovering] = useState(false)

  const startAnimation = useCallback(() => {
    setIsAnimating(true)
    const finalText = text.split('')
    let currentIteration = 0
    const maxIterations = 8

    const interval = setInterval(() => {
      setDisplayText((prev) =>
        prev.map((char, index) => {
          // If we've reached the final character, keep it
          if (currentIteration >= maxIterations) {
            return finalText[index]
          }

          // Randomly flip through characters before settling
          if (currentIteration > index * 0.5) {
            // Start settling characters progressively
            const progress = (currentIteration - index * 0.5) / (maxIterations - index * 0.5)
            if (progress > 0.7) {
              return finalText[index]
            }
          }

          // Random character while flipping
          return CHARACTERS[Math.floor(Math.random() * CHARACTERS.length)]
        })
      )

      currentIteration++

      if (currentIteration > maxIterations) {
        clearInterval(interval)
        setDisplayText(finalText)
        setIsAnimating(false)
      }
    }, 60) // Faster animation on hover

    return interval
  }, [text])

  // Initial animation on mount
  useEffect(() => {
    if (!isVisible) return

    const interval = startAnimation()
    return () => clearInterval(interval)
  }, [isVisible, startAnimation])

  // Hover animation
  const handleMouseEnter = () => {
    setIsHovering(true)
    if (!isAnimating) {
      startAnimation()
    }
  }

  const handleMouseLeave = () => {
    setIsHovering(false)
  }

  return (
    <div 
      className="inline-flex transition-all duration-300"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        transform: isHovering ? 'scale(1.01)' : 'scale(1)',
      }}
    >
      {displayText.map((char, index) => (
        <div
          key={index}
          className={`relative inline-block ${char === ' ' ? 'w-4 md:w-6 lg:w-8' : ''} ${
            isAnimating ? 'animating' : ''
          }`}
          style={{
            animation: isAnimating ? `flip 0.5s ease-out ${index * 0.04}s` : 'none',
          }}
        >
          <span className="inline-block">
            {char === ' ' ? '\u00A0' : char}
          </span>
        </div>
      ))}
      <style jsx>{`
        @keyframes flip {
          0% {
            transform: rotateX(90deg);
            opacity: 0;
          }
          50% {
            transform: rotateX(-10deg);
          }
          100% {
            transform: rotateX(0deg);
            opacity: 1;
          }
        }
      `}</style>
    </div>
  )
}

