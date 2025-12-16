'use client'

import { useState, useEffect } from 'react'

interface FlipTextProps {
  text: string
  isVisible: boolean
}

const CHARACTERS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 '

export const FlipText = ({ text, isVisible }: FlipTextProps) => {
  const [displayText, setDisplayText] = useState(text.split('').map(() => ' '))

  useEffect(() => {
    if (!isVisible) return

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
      }
    }, 80)

    return () => clearInterval(interval)
  }, [isVisible, text])

  return (
    <div className="inline-flex">
      {displayText.map((char, index) => (
        <div
          key={index}
          className={`relative inline-block ${char === ' ' ? 'w-4 md:w-6 lg:w-8' : ''}`}
          style={{
            animation: isVisible ? `flip 0.6s ease-out ${index * 0.05}s` : 'none',
          }}
        >
          <span className="inline-block">{char === ' ' ? '\u00A0' : char}</span>
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

