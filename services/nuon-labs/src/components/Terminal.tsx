'use client'

import { useState, useEffect, useRef } from 'react'

declare global {
  interface Window {
    UnicornStudio?: {
      isInitialized: boolean
      init: () => void
    }
  }
}

interface Command {
  input: string
  output: string[]
  isIntro?: boolean
}

const COMMANDS: Record<string, string[] | ((args: string[]) => string[])> = {
  'help': [
    '',
    '  Available commands:',
    '',
    '    help              Show this help message',
    '    list              List all experiments',
    '    open <project>    Open an experiment',
    '    status <project>  Check project status',
    '    about             Learn about Nuon Labs',
    '    clear             Clear the terminal',
    '',
  ],
  'about': [
    '',
    '  Nuon Labs - Experimental BYOC Playground',
    '',
    '  This is where we build and test cutting-edge features',
    '  before they graduate to production.',
    '',
    '  All experiments are in alpha/beta.',
    '',
  ],
  'list': [
    '',
    '  Active Experiments:',
    '',
    '    ● customer-dashboard   [alpha]',
    '      Purpose-built dashboard for customers',
    '',
    '  Use "open <project>" to try one out!',
    '',
  ],
  'open': (args: string[]) => {
    const project = args[0]?.toLowerCase()
    const projects: Record<string, { url: string; status: string }> = {
      'customer-dashboard': {
        url: 'https://vendor.inl0qjpbg8hn5e25ebmcjzmwh2.nuon.run/',
        status: 'alpha',
      },
    }

    if (!project) {
      return ['', '  Usage: open <project>', '  Try: open customer-dashboard', '']
    }

    if (projects[project]) {
      setTimeout(() => {
        window.open(projects[project].url, '_blank')
      }, 100)
      return [
        '',
        `  ✓ Opening ${project}...`,
        `    Status: ${projects[project].status}`,
        '',
      ]
    }

    return [
      '',
      `  ✗ Project "${project}" not found.`,
      '  Use "list" to see available experiments.',
      '',
    ]
  },
  'status': (args: string[]) => {
    const project = args[0]?.toLowerCase()
    const statuses: Record<string, string[]> = {
      'customer-dashboard': [
        '',
        '  customer-dashboard',
        '    Status: alpha',
        '    Updated: 2 days ago',
        '    Features: Install management, update approvals',
        '',
      ],
    }

    if (!project) {
      return ['', '  Usage: status <project>', '  Try: status customer-dashboard', '']
    }

    return statuses[project] || ['', `  ✗ Project "${project}" not found.`, '']
  },
}

const ALL_COMMANDS = ['help', 'list', 'open', 'status', 'about', 'clear']
const PROJECTS = ['customer-dashboard']

export const Terminal = () => {
  const [isVisible, setIsVisible] = useState(false)
  const [history, setHistory] = useState<Command[]>([])
  const [commandHistory, setCommandHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [input, setInput] = useState('')
  const [suggestion, setSuggestion] = useState('')
  const [showIntro, setShowIntro] = useState(true)
  const [isReady, setIsReady] = useState(false)
  const [isHovering, setIsHovering] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const terminalRef = useRef<HTMLDivElement>(null)
  const unicornLoaded = useRef(false)

  useEffect(() => {
    // Initialize terminal visibility and load WebGL
    setIsVisible(true)
    
    // Load Unicorn Studio script
    if (!unicornLoaded.current && !window.UnicornStudio) {
      const script = document.createElement('script')
      script.src = 'https://cdn.jsdelivr.net/gh/hiunicornstudio/unicornstudio.js@v1.5.3/dist/unicornStudio.umd.js'
      script.onload = () => {
        if (window.UnicornStudio && !window.UnicornStudio.isInitialized) {
          window.UnicornStudio.init()
          window.UnicornStudio.isInitialized = true
        }
      }
      document.head.appendChild(script)
      unicornLoaded.current = true
    } else if (window.UnicornStudio && !window.UnicornStudio.isInitialized) {
      window.UnicornStudio.init()
      window.UnicornStudio.isInitialized = true
    }

    // Show welcome message after intro animation
    setTimeout(() => {
      setShowIntro(false)
      setIsReady(true)
    }, 4000) // Show intro for 4 seconds
  }, [])

  // Auto-scroll to bottom when history changes
  useEffect(() => {
    if (terminalRef.current) {
      // Use requestAnimationFrame to ensure DOM is updated
      requestAnimationFrame(() => {
        if (terminalRef.current) {
          terminalRef.current.scrollTop = terminalRef.current.scrollHeight
        }
      })
    }
  }, [history])

  // Auto-focus input when terminal becomes ready (without scrolling page)
  useEffect(() => {
    if (isReady && inputRef.current) {
      inputRef.current.focus({ preventScroll: true })
    }
  }, [isReady])

  // Autocomplete logic
  useEffect(() => {
    if (!input) {
      setSuggestion('')
      return
    }

    const parts = input.split(' ')
    const cmd = parts[0].toLowerCase()
    const arg = parts[1]?.toLowerCase() || ''

    // Command autocomplete
    if (parts.length === 1) {
      const match = ALL_COMMANDS.find(c => c.startsWith(cmd) && c !== cmd)
      if (match) {
        setSuggestion(match.slice(cmd.length))
      } else {
        setSuggestion('')
      }
    }
    // Project argument autocomplete
    else if (parts.length === 2 && (cmd === 'open' || cmd === 'status')) {
      const match = PROJECTS.find(p => p.startsWith(arg) && p !== arg)
      if (match) {
        setSuggestion(match.slice(arg.length))
      } else {
        setSuggestion('')
      }
    } else {
      setSuggestion('')
    }
  }, [input])

  const handleCommand = (cmd: string) => {
    const trimmedCmd = cmd.trim()
    const parts = trimmedCmd.split(' ')
    const command = parts[0].toLowerCase()
    const args = parts.slice(1)
    
    if (command === 'clear') {
      setHistory([])
      return
    }

    let output: string[]
    if (trimmedCmd === '') {
      output = []
    } else if (COMMANDS[command]) {
      const commandHandler = COMMANDS[command]
      if (typeof commandHandler === 'function') {
        output = commandHandler(args)
      } else {
        output = commandHandler
      }
    } else {
      output = ['', `  ✗ Unknown command: ${command}`, '  Type help for available commands', '']
    }

    setHistory([...history, { input: cmd, output }])
    if (cmd.trim()) {
      setCommandHistory(prev => [...prev, cmd])
    }
    setHistoryIndex(-1)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    // Tab completion
    if (e.key === 'Tab') {
      e.preventDefault()
      if (suggestion) {
        setInput(input + suggestion)
        setSuggestion('')
      }
      return
    }

    // Up arrow - previous command
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (commandHistory.length > 0) {
        const newIndex = historyIndex === -1 
          ? commandHistory.length - 1 
          : Math.max(0, historyIndex - 1)
        setHistoryIndex(newIndex)
        setInput(commandHistory[newIndex])
        setSuggestion('')
      }
      return
    }

    // Down arrow - next command
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex !== -1) {
        const newIndex = historyIndex + 1
        if (newIndex >= commandHistory.length) {
          setHistoryIndex(-1)
          setInput('')
        } else {
          setHistoryIndex(newIndex)
          setInput(commandHistory[newIndex])
        }
        setSuggestion('')
      }
      return
    }

    // Escape - clear input
    if (e.key === 'Escape') {
      setInput('')
      setSuggestion('')
      setHistoryIndex(-1)
      return
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (input.trim()) {
      handleCommand(input)
      setInput('')
    }
  }

  const handleTerminalClick = () => {
    inputRef.current?.focus()
  }

  return (
    <div
      className={`w-full sm:w-[90%] lg:w-[80%] mx-auto transition-all duration-1000 relative z-10 ${
        isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'
      }`}
    >
      <div
        className="bg-[#111111] border border-[#2a2a2a] rounded-xl overflow-hidden shadow-2xl transition-all duration-500"
        style={{
          boxShadow: isHovering 
            ? '0 25px 50px -12px rgba(59, 130, 246, 0.4), 0 10px 30px -5px rgba(59, 130, 246, 0.3), 0 0 60px rgba(59, 130, 246, 0.2)'
            : '0 25px 50px -12px rgba(0, 0, 0, 0.5), 0 10px 20px -5px rgba(0, 0, 0, 0.3)',
          borderColor: isHovering ? 'rgba(59, 130, 246, 0.3)' : '#2a2a2a',
        }}
        onClick={handleTerminalClick}
        onMouseEnter={() => setIsHovering(true)}
        onMouseLeave={() => setIsHovering(false)}
      >
        {/* Terminal Header */}
        <div className="flex items-center gap-2 px-4 py-2.5 bg-[#191919] border-b border-[#2a2a2a]">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
            <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
            <div className="w-3 h-3 rounded-full bg-[#28c840]" />
          </div>
          <div className="flex-1 text-center text-xs text-[#666666]">
            nuon-labs
          </div>
        </div>

        {/* Terminal Body - with WebGL overlay */}
        <div className="relative h-[400px] sm:h-[500px] lg:h-[550px]">
          {/* WebGL Animation - Overlay on right 50% (hidden on mobile) */}
          <div 
            className="hidden md:block absolute top-0 right-0 w-1/2 h-full pointer-events-none z-0"
            style={{ 
              maskImage: 'linear-gradient(to right, transparent 0%, black 30%)',
              WebkitMaskImage: 'linear-gradient(to right, transparent 0%, black 30%)',
            }}
          >
            <div 
              data-us-project="up8EyNwsmxx9SXw6iC5g" 
              style={{ width: '100%', height: '100%' }}
            />
          </div>

          {/* Terminal Content */}
          <div
            ref={terminalRef}
            className="relative z-10 p-4 sm:p-6 h-full overflow-y-auto text-xs sm:text-sm flex flex-col custom-scrollbar"
          >
            {/* Spacer to push content to bottom */}
            <div className="flex-1 min-h-0"></div>
            
            {/* Terminal Content - MOTD + Command History */}
            <div>
              {/* Welcome Message (MOTD) */}
              <div className="mb-4 sm:mb-6">
                <div className="text-[#e5e5e5] text-sm font-mono mb-2">
                  Nuon Labs
                </div>
                <div className="text-[#808080] text-xs mb-3">
                  Where BYOC ideas take shape.
                </div>
                <div className="text-[#666666] text-xs space-y-1 mb-4">
                  <div>Type <span className="text-[#f97316]">help</span> for available commands</div>
                  <div>Type <span className="text-[#f97316]">list</span> to see active experiments</div>
                </div>
              </div>

              {/* Command History */}
              {history.map((cmd, i) => (
                <div key={i} className="mb-3">
                  {cmd.input && (
                    <div className="flex gap-2 items-center text-sm">
                      <span className="text-[#f97316]">❯</span>
                      <span className="text-[#e5e5e5]">{cmd.input}</span>
                    </div>
                  )}
                  {cmd.output.length > 0 && (
                    <div className="text-[#a3a3a3] whitespace-pre pl-4 text-sm">
                      {cmd.output.map((line, j) => (
                        <div key={j}>{line}</div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Fixed Input at Bottom */}
        <div className="px-4 sm:px-5 py-3 sm:py-5 border-t border-[#2a2a2a] bg-[#0d0d0d]">
          <form onSubmit={handleSubmit} className="flex gap-2 items-center text-sm sm:text-base">
            <span className="text-[#f97316]">❯</span>
            <div className="flex-1 relative py-1">
              <span className="absolute left-0 top-1 pointer-events-none whitespace-pre text-sm sm:text-base">
                <span className="text-[#e5e5e5]">{input}</span>
                <span className="text-[#525252]">{suggestion}</span>
              </span>
              <input
                ref={inputRef}
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full bg-transparent outline-none text-transparent caret-[#f97316] text-sm sm:text-base"
                style={{ caretColor: '#f97316' }}
                spellCheck={false}
                autoComplete="off"
              />
            </div>
            {suggestion && (
              <span className="hidden sm:inline text-[#525252] text-xs">tab</span>
            )}
          </form>
          <div className="hidden sm:flex gap-4 mt-2 text-[11px] text-[#525252]">
            <span><span className="text-[#666666]">tab</span> complete</span>
            <span><span className="text-[#666666]">↑↓</span> history</span>
            <span><span className="text-[#666666]">esc</span> clear</span>
          </div>
        </div>
      </div>
    </div>
  )
}
