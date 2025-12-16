export const Footer = () => {
  return (
    <footer className="mt-auto pt-40 pb-8">
      <div className="w-[80%] mx-auto">
        {/* Divider with gradient */}
        <div className="h-px bg-gradient-to-r from-transparent via-[#2a2a2a] to-transparent mb-8" />
        
        {/* Footer row */}
        <div className="flex items-center justify-center gap-6 text-[11px] text-[#525252] uppercase tracking-wider">
          <span>© {new Date().getFullYear()}</span>
          <span className="text-[#2a2a2a]">/</span>
          <a
            href="https://nuon.co"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-white transition-colors"
          >
            Nuon
          </a>
          <span className="text-[#2a2a2a]">/</span>
          <span className="text-[#404040]">Labs</span>
        </div>
      </div>
    </footer>
  )
}
