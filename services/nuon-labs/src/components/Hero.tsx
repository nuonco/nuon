'use client'

import { useEffect, useState } from 'react'
import { Terminal } from './Terminal'

const PROJECTS = [
  {
    date: '2024.12.15',
    name: 'Customer Dashboard',
    description: 'Purpose-built dashboard for customers to install apps and approve updates',
    status: 'alpha',
    url: 'https://vendor.inl0qjpbg8hn5e25ebmcjzmwh2.nuon.run/',
  },
]

export const Hero = () => {
  const [heroVisible, setHeroVisible] = useState(false)
  const [terminalVisible, setTerminalVisible] = useState(false)

  useEffect(() => {
    // Show hero text and terminal together after intro animation completes
    const handleIntroComplete = () => {
      setHeroVisible(true)
      setTerminalVisible(true)
    }

    window.addEventListener('intro-complete', handleIntroComplete)
    
    return () => {
      window.removeEventListener('intro-complete', handleIntroComplete)
    }
  }, [])

  return (
    <>
      {/* Hero Section - Full viewport height with centered title */}
      <div className="min-h-screen flex flex-col items-center justify-center px-6">
        <div 
          className={`text-center transition-all duration-1000 ${
            heroVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
          }`}
        >
          <h1 
            className="text-6xl md:text-7xl lg:text-8xl text-white tracking-tight font-bold"
            style={{ fontFamily: 'var(--font-ibm-plex-mono)' }}
          >
            Nuon Labs
          </h1>
          <p className="text-base text-white/70 mt-3">
            Experimental BYOC Playground
          </p>
        </div>
      </div>

      {/* Terminal Section - only visible after scroll */}
      <div 
        className={`px-6 pb-32 -mt-48 transition-all duration-1000 ${
          terminalVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-12'
        }`}
      >
        <Terminal />
      </div>

      {/* Projects List Section - shows after terminal */}
      <div 
        className={`w-[80%] mx-auto px-6 pb-20 mt-8 transition-all duration-1000 delay-300 ${
          terminalVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
        }`}
      >
        {/* Title */}
        <div className="mb-8">
          <h2 className="text-4xl md:text-5xl text-[#e5e5e5] font-bold tracking-tight" style={{ fontFamily: 'var(--font-ibm-plex-mono)' }}>
            Experiments
            <sup className="text-base text-[#666666] ml-1">({PROJECTS.length})</sup>
          </h2>
        </div>

        {/* Table Header */}
        <div className="grid grid-cols-[140px_1fr_100px_40px] gap-4 pb-3 border-b border-[#2a2a2a] text-[11px] text-[#666666] uppercase tracking-wider">
          <div>/ Date</div>
          <div>/ Name</div>
          <div>/ Status</div>
          <div></div>
        </div>

        {/* Table Rows */}
        <div>
          {PROJECTS.map((project) => (
            <a 
              key={project.name} 
              href={project.url}
              target="_blank"
              rel="noopener noreferrer"
              className="grid grid-cols-[140px_1fr_100px_40px] gap-4 py-4 border-b border-[#2a2a2a] hover:bg-[#111111] transition-colors cursor-pointer group items-center"
            >
              <div className="flex items-center gap-2 text-sm text-[#808080]">
                <span className="text-[#f97316]">■</span>
                {project.date}
              </div>
              <div className="text-[#e5e5e5] text-base">
                {project.name}
              </div>
              <div>
                <span className="text-[11px] px-2 py-1 border border-[#2a2a2a] text-[#808080] uppercase tracking-wider">
                  {project.status}
                </span>
              </div>
              <div className="text-[#525252] group-hover:text-[#e5e5e5] transition-colors text-xl">
                +
              </div>
            </a>
          ))}
        </div>
      </div>
    </>
  )
}
