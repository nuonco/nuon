export const Logo = () => {
  return (
    <a href="/">
      <span className="sr-only">Nuon</span>
      <img
        className="w-auto h-7 relative block dark:hidden"
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/light.svg"
        alt="light logo"
      />
      <img
        className="w-auto h-7 relative hidden dark:block"
        src="https://mintlify.s3-us-west-1.amazonaws.com/nuoninc/logo/dark.svg"
        alt="dark logo"
      />
    </a>
  )
}
