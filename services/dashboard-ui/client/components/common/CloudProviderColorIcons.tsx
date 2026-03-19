import { FaAws } from 'react-icons/fa'
import { SiGooglecloud } from 'react-icons/si'
import { VscAzure } from 'react-icons/vsc'

interface IColorIcon {
  size?: number | string
  className?: string
}

export const AWSColor = ({ size = 16, className }: IColorIcon) => (
  <FaAws size={size} className={className} style={{ color: '#FF9900' }} />
)

export const AzureColor = ({ size = 16, className }: IColorIcon) => (
  <VscAzure size={size} className={className} style={{ color: '#0078D4' }} />
)

export const GCPColor = ({ size = 16, className }: IColorIcon) => (
  <SiGooglecloud size={size} className={className} style={{ color: '#4285F4' }} />
)
