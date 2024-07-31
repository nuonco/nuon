import { Button } from "@material-tailwind/react";

export const InstallButton = ({ install }) => {
  const loading = install.status === "provisioning";

  let label = "Create Install";
  if (install && install.id && install.id.length > 0) {
    label = "Reprovision Install";
  }
  if (loading) {
    label = "Provisioning";
  }

  return (
    <Button
      loading={loading}
      type="submit"
      className={`
          text-sm text-button-text-color
          bg-button-bg-color hover:bg-button-bg-hover focus:bg-button-bg-focus active:bg-button-bg-active
          !shadow-button-shadow !dark:shadow-button-shadow-dark
          !border-button-border-color border-button-border-width rounded-button-radius
          px-4 py-1.5 mr-1`}
    >
      {label}
    </Button>
  );
};
