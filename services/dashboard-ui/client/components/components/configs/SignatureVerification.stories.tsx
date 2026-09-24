export default {
  title: 'Components/Configs/SignatureVerification',
}

import { SignatureVerification } from './SignatureVerification'

const publicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEqrQ0Kx0hZ7Vv0Yy5bqz1s6dGm2nF
9Xk0mQp3rJ8tLw1cV6hYy4sD2oN7bQ0fR5uT8eK1jZ3xM9vA6pC4nW0gHQ==
-----END PUBLIC KEY-----`

export const Keyless = () => (
  <SignatureVerification
    verification={{
      require_signature: true,
      authorities: [
        {
          type: 'keyless',
          issuer: 'https://token.actions.githubusercontent.com',
          subject:
            'https://github.com/acme/api/.github/workflows/release.yaml@refs/heads/main',
        },
      ],
    }}
  />
)

export const KeylessWithSubjectPattern = () => (
  <SignatureVerification
    verification={{
      require_signature: true,
      authorities: [
        {
          type: 'keyless',
          issuer: 'https://token.actions.githubusercontent.com',
          subject_regexp:
            '^https://github\\.com/acme/api/\\.github/workflows/release\\.yaml@refs/tags/v[0-9]+\\.[0-9]+\\.[0-9]+$',
        },
      ],
    }}
  />
)

export const PublicKey = () => (
  <SignatureVerification
    verification={{
      require_signature: true,
      authorities: [{ type: 'public_key', public_key: publicKey }],
    }}
  />
)

export const MultipleAuthorities = () => (
  <SignatureVerification
    verification={{
      require_signature: true,
      authorities: [
        { type: 'public_key', public_key: publicKey },
        {
          type: 'keyless',
          issuer: 'https://token.actions.githubusercontent.com',
          subject:
            'https://github.com/acme/api/.github/workflows/release.yaml@refs/heads/main',
        },
      ],
    }}
  />
)

export const NotRequired = () => (
  <SignatureVerification
    verification={{
      require_signature: false,
      authorities: [
        {
          type: 'keyless',
          issuer: 'https://token.actions.githubusercontent.com',
          subject:
            'https://github.com/acme/api/.github/workflows/release.yaml@refs/heads/main',
        },
      ],
    }}
  />
)

export const NoAuthorities = () => (
  <SignatureVerification verification={{ require_signature: true }} />
)

export const NoVerificationBlock = () => (
  <SignatureVerification verification={undefined} />
)
