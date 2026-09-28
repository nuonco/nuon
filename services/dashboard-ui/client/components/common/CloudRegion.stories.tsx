import { CloudRegion } from './CloudRegion'

export default { title: 'UI / Cloud region' }

export const AWS = () => <CloudRegion platform="aws" region="us-east-1" />

export const Azure = () => <CloudRegion platform="azure" location="eastus" />

export const GCP = () => <CloudRegion platform="gcp" region="us-central1" />

export const Unknown = () => <CloudRegion platform="aws" region="invalid-region" />
