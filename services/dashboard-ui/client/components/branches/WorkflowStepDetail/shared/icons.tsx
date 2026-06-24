export const DetailStatusIcon = ({ status }: { status?: string }) => {
  if (status === 'success' || status === 'succeeded') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-green-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M2.5 6.5L5.5 9.5L10.5 4" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }
  if (status === 'error') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-red-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M4 4L9 9M9 4L4 9" stroke="white" strokeWidth="1.8" strokeLinecap="round" />
        </svg>
      </div>
    )
  }
  if (status === 'in-progress') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
        <svg className="animate-spin" width="16" height="16" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke="rgba(255,255,255,0.3)" strokeWidth="2" />
          <path d="M8 2 A6 6 0 0 1 14 8" stroke="white" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </div>
    )
  }
  return (
    <div
      className="w-[26px] h-[26px] rounded-full flex items-center justify-center shrink-0"
      style={{ boxShadow: 'inset 0 0 0 1.5px rgba(150,150,170,0.35)' }}
    >
      <div className="w-[5px] h-[5px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

export const InstallStatusIcon = ({ status }: { status?: string }) => {
  if (status === 'success' || status === 'deployed') {
    return (
      <div className="w-[17px] h-[17px] rounded-full border-2 border-green-500 flex items-center justify-center shrink-0">
        <div className="w-[5px] h-[5px] rounded-full bg-green-500" />
      </div>
    )
  }
  if (status === 'in-progress') {
    return (
      <div className="w-[17px] h-[17px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
        <svg className="animate-spin" width="11" height="11" viewBox="0 0 11 11" fill="none">
          <circle cx="5.5" cy="5.5" r="4" stroke="rgba(255,255,255,0.3)" strokeWidth="1.5" />
          <path d="M5.5 1.5 A4 4 0 0 1 9.5 5.5" stroke="white" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
    )
  }
  if (status === 'error') {
    return (
      <div className="w-[17px] h-[17px] rounded-full border-2 border-red-500 flex items-center justify-center shrink-0">
        <div className="w-[5px] h-[5px] rounded-full bg-red-500" />
      </div>
    )
  }
  return (
    <div
      className="w-[17px] h-[17px] rounded-full flex items-center justify-center shrink-0"
      style={{ boxShadow: 'inset 0 0 0 1.5px rgba(150,150,170,0.35)' }}
    >
      <div className="w-[4px] h-[4px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

export const DiffMarker = ({ op }: { op?: string }) => {
  if (op === 'add') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-green-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-green-600 dark:text-green-400 leading-none">+</span>
      </div>
    )
  }
  if (op === 'remove') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-red-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-red-600 dark:text-red-400 leading-none">−</span>
      </div>
    )
  }
  if (op === 'change') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-yellow-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-yellow-600 dark:text-yellow-400 leading-none">~</span>
      </div>
    )
  }
  return null
}
