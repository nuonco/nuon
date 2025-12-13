'use client'

import { Text } from '@/components/common/Text'

export const Header = () => {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 px-6 py-4">
      <div className="max-w-6xl mx-auto flex items-center">
        <div className="flex items-center gap-3">
          <NuonLogo />
          <div className="h-6 w-px bg-white/20" />
          <Text
            variant="body"
            weight="strong"
            className="text-primary-400 uppercase tracking-wider text-xs"
          >
            Labs
          </Text>
        </div>
      </div>
    </header>
  )
}

const NuonLogo = () => (
  <svg
    width="32"
    height="32"
    viewBox="0 0 1080 1080"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="h-6 w-6"
  >
    <defs>
      <linearGradient
        id="nuon-gradient"
        x1="359.74"
        y1="979.23"
        x2="717.77"
        y2="104.97"
        gradientUnits="userSpaceOnUse"
      >
        <stop offset="0" stopColor="#F72585" />
        <stop offset="0.53" stopColor="#3A00FF" />
        <stop offset="1" stopColor="#4CC9F0" />
      </linearGradient>
    </defs>
    <path
      d="M704.25,72.41L519.57,175.78v167.36l-149.5-83.72h-0.09l-178.2,99.74v548.81l178.12,99.74h0.09l191.24-107.09V740.61l143.02,80.01l184.68-103.37V175.78L704.25,72.41z M233.52,382.52l136.38-76.29h0.09l149.5,83.64v280.64L233.52,510.5V382.52z M519.48,877.25L369.9,960.89L233.52,884.6V557.23l285.96,160.01V877.25z M847.19,693.79L704.25,773.8l-142.94-79.92V413.24l285.96,160.01v120.54H847.19z M847.19,526.52L561.22,366.51V199.15l143.02-80.01l142.94,80.01V526.52z"
      fill="url(#nuon-gradient)"
    />
  </svg>
)
