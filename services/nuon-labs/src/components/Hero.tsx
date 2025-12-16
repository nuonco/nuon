'use client'

import { useEffect, useState } from 'react'
import { Terminal } from './Terminal'
import { FlipText } from './FlipText'

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
      <div className="min-h-screen flex flex-col items-center justify-center px-4 sm:px-6">
        <div 
          className={`text-center transition-all duration-1000 ${
            heroVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
          }`}
        >
          <h1 
            className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl xl:text-8xl text-white tracking-tight font-bold"
          >
            <FlipText text="Nuon Labs" isVisible={heroVisible} />
          </h1>
          <p className="text-sm sm:text-base text-white/70 mt-2 sm:mt-3">
          Nuon Labs is where BYOC experiments take shape.
          </p>
        </div>
      </div>

      {/* Terminal Section - hidden on mobile, visible on md+ */}
      <div 
        className={`hidden md:block px-4 sm:px-6 pb-32 -mt-32 sm:-mt-48 lg:-mt-64 transition-all duration-1000 ${
          terminalVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-12'
        }`}
      >
        <Terminal />
      </div>

      {/* Projects List Section - shows after terminal */}
      <div 
        className={`w-full sm:w-[90%] lg:w-[80%] mx-auto px-4 sm:px-6 pb-20 mt-8 md:mt-8 -mt-32 transition-all duration-1000 delay-300 ${
          terminalVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
        }`}
      >
        {/* Title */}
        <div className="mb-6 sm:mb-8">
          <h2 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl text-[#e5e5e5] font-bold tracking-tight">
            Experiments
            <sup className="text-sm sm:text-base text-[#666666] ml-1">({PROJECTS.length})</sup>
          </h2>
        </div>

        {/* Mobile: Modern Card Layout */}
        <div className="block md:hidden space-y-3">
          {PROJECTS.map((project) => (
            <a 
              key={project.name} 
              href={project.url}
              target="_blank"
              rel="noopener noreferrer"
              className="block group"
            >
              <div className="bg-[#111111] border border-[#2a2a2a] rounded-lg p-5 hover:border-[#f97316]/30 transition-all">
                {/* Header row */}
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="text-[#e5e5e5] text-lg font-semibold mb-1 group-hover:text-white transition-colors">
                      {project.name}
                    </div>
                    <div className="text-[#666666] text-sm leading-relaxed">
                      {project.description}
                    </div>
                  </div>
                </div>
                
                {/* Footer row */}
                <div className="flex items-center justify-between pt-3 border-t border-[#1a1a1a]">
                  <div className="flex items-center gap-2 text-xs text-[#808080]">
                    <span className="text-[#f97316]">●</span>
                    <span>{project.date}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-[10px] px-2.5 py-1 bg-[#0d0d0d] border border-[#2a2a2a] text-[#808080] uppercase tracking-wider rounded">
                      {project.status}
                    </span>
                    <span className="text-[#525252] group-hover:text-[#f97316] transition-colors">→</span>
                  </div>
                </div>
              </div>
            </a>
          ))}
        </div>

        {/* Desktop: Table Layout */}
        <div className="hidden md:block">
          {/* Table Header */}
          <div className="grid grid-cols-[120px_1fr_90px_40px] lg:grid-cols-[140px_1fr_100px_40px] gap-4 pb-3 border-b border-[#2a2a2a] text-[11px] text-[#666666] uppercase tracking-wider">
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
                className="grid grid-cols-[120px_1fr_90px_40px] lg:grid-cols-[140px_1fr_100px_40px] gap-4 py-4 pr-6 border-b border-[#2a2a2a] hover:bg-[#111111] transition-colors cursor-pointer group items-center"
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
                <div className="flex items-center gap-1.5 text-[#525252] group-hover:text-[#006CFF] transition-colors text-xs uppercase tracking-wider overflow-visible">
                  <span>View</span>
                  <svg 
                    width="14" 
                    height="14" 
                    viewBox="0 0 14 14" 
                    fill="none" 
                    className="flex-shrink-0"
                    style={{ minWidth: '14px', minHeight: '14px' }}
                  >
                    <path 
                      d="M4 10L10 4M10 4H5M10 4V9" 
                      stroke="currentColor" 
                      strokeWidth="1.5" 
                      strokeLinecap="round" 
                      strokeLinejoin="round"
                    />
                  </svg>
                </div>
              </a>
            ))}
          </div>
        </div>
      </div>
    </>
  )
}
