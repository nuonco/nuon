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
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 84 32"
    className="h-6 w-auto"
  >
    <path
      d="M16.992 0 10.87 3.537v5.726L5.912 6.398h-.003L0 9.811v18.777L5.906 32h.003l6.34-3.664v-5.474l4.743 2.737 6.124-3.537V3.537L16.992 0ZM1.384 10.61 5.906 8h.003l4.957 2.862v9.601L1.384 14.99v-4.38Zm9.482 16.927-4.96 2.861-4.522-2.61v-11.2l9.482 5.474v5.475Zm10.866-6.277-4.74 2.737-4.74-2.734V11.66l9.482 5.475v4.124h-.002Zm0-5.723-9.482-5.475V4.336L16.992 1.6l4.74 2.737v11.2Z"
      fill="url(#nuon-logo-gradient)"
    />
    <g transform="translate(4, -5)">
      <path
        d="M40.824 19.172v6.44h-2.78v-5.784c0-1.695-.79-2.419-2.146-2.419-1.424 0-2.44.837-2.44 3.096v5.108H30.7v-10.44h2.734v1.581h.045c.7-1.198 1.875-1.92 3.412-1.92 2.26 0 3.932 1.265 3.932 4.338Zm1.799 2.396v-6.396h2.847v5.898c0 1.628.768 2.328 1.966 2.328 1.378 0 2.305-.904 2.305-3.209v-5.017h2.78v10.44h-2.735v-1.604h-.045c-.633 1.13-1.627 1.944-3.254 1.944-2.124 0-3.864-1.288-3.864-4.384Zm22.867-1.176c0 3.277-2.463 5.56-5.784 5.56-3.322 0-5.785-2.283-5.785-5.56 0-3.276 2.463-5.559 5.785-5.559 3.321 0 5.784 2.283 5.784 5.56Zm-8.722 0c0 1.786 1.243 2.938 2.938 2.938 1.694 0 2.914-1.152 2.914-2.938 0-1.785-1.22-2.937-2.914-2.937-1.695 0-2.938 1.152-2.938 2.938ZM77 19.172v6.44h-2.78v-5.784c0-1.695-.79-2.419-2.146-2.419-1.424 0-2.44.837-2.44 3.096v5.108h-2.757v-10.44h2.734v1.581h.045c.7-1.198 1.876-1.92 3.412-1.92 2.26 0 3.932 1.265 3.932 4.338Z"
        fill="currentColor"
      />
    </g>
    <defs>
      <linearGradient
        id="nuon-logo-gradient"
        x1="0"
        y1="28.672"
        x2="12.523"
        y2="-.964"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="#F72585" />
        <stop offset=".53" stopColor="#3A00FF" />
        <stop offset="1" stopColor="#4CC9F0" />
      </linearGradient>
    </defs>
  </svg>
)
