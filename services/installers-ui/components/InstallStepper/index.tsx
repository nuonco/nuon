"use client";

import React, { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

import { Stepper, Step } from "../Stepper";
import { Typography } from "@material-tailwind/react";

import { NavArrows } from "./NavArrows";
import { CompanyContent } from "./CompanyContent";
import { CloudAccountContent } from "./CloudAccountContent";
import { GroupContent } from "./GroupContent";
import { InstallStatusContent } from "./InstallStatusContent";
import { ErrorAlert } from "./ErrorAlert";

const InstallStepper = ({
  app,
  existingInstall,
  installer,
  searchParams,
  regions,
  createInstall,
  getInstall,
  redeployInstall,
}) => {
  // track state of install
  const [install, setInstall] = React.useState(
    existingInstall
      ? existingInstall
      : {
          id: searchParams.install_id || "",
          status: "not created",
          status_description: "install has not been created yet",
          install_components: [],
          install_inputs: [],
        },
  );

  // TODO: break this out into a testable function
  const _install_inputs = install.install_inputs[0] || [{ values: {} }];
  const install_input_values = _install_inputs.values;
  const input_groups = app.input_config.input_groups || [];

  const [activeStep, setActiveStep] = React.useState(
    existingInstall ? input_groups.length + 2 : 0,
  );
  const [isLastStep, setIsLastStep] = React.useState(false);
  const [isFirstStep, setIsFirstStep] = React.useState(false);

  const handleNext = () => !isLastStep && setActiveStep((cur) => cur + 1);
  const handlePrev = () => !isFirstStep && setActiveStep((cur) => cur - 1);
  const router = useRouter();

  const [error, setError] = React.useState({
    description: "",
    error: "",
    user_error: false,
  });

  // create or update install when form is submitted
  const formAction = async (event) => {
    event.preventDefault();

    if (event.key == "Enter") {
      // this should have been caught by onKeyUp but we include it here also
      // TODO: trigger required form field validation
      handleNext();
    }

    const formData = new FormData(event.target);
    let installID = "";
    if (install.id === "") {
      // if we haven't created the install yet, create it
      const res = await createInstall(app, formData, app.org_id);
      if (res.error) {
        setError(res);
        return;
      }
      installID = res.id;
      // navigate to the sub-route
      router.push(`/${app.name}/${res.id}`);
    } else {
      // if we've already created the install, redeploy it
      const reproRes = await redeployInstall(
        install.id,
        app,
        formData,
        app.org_id,
      );
      if (reproRes.error) {
        setError(reproRes);
        return;
      }
    }

    // by this point we should have an install ID
    // fetch the install so we can render the status
    const res = await getInstall(installID, app.org_id);
    if (res.error) {
      setError(res);
      return;
    }

    setInstall(res);

    setError({
      description: "",
      error: "",
      user_error: false,
    });
  };

  // poll for install status once we have an ID
  const useInterval = (callback, delay) => {
    const savedCallback = useRef();

    useEffect(() => {
      savedCallback.current = callback;
    }, [callback]);

    useEffect(() => {
      function tick() {
        // `savedCallback.current()` works fine in dev, but fails the build.
        // Basically, Typescript can't know if savedCallback.current is defined in the context of this closure,
        // so we need to assert that it's a function in order for Typescript to accept it.
        (savedCallback.current as () => {})();
      }
      if (delay !== null) {
        const id = setInterval(tick, delay);
        return () => clearInterval(id);
      }
    }, [delay]);
  };
  useInterval(async () => {
    if (install.id !== "") {
      const res = await getInstall(install.id, app.org_id);
      if (res.error) {
        setError(res);
        return;
      }

      setInstall(res);
      setError({
        description: "",
        error: "",
        user_error: false,
      });
    }
  }, 1000 * 10);

  const stepContent = input_groups.map((group, idx) => (
    <GroupContent
      key={idx}
      group={group}
      install_input_values={install_input_values}
      idx={idx}
      setActiveStep={() => setActiveStep(idx + 1)}
      activeStep={activeStep}
    />
  ));

  const steps = input_groups.map((group, idx) => (
    <Step
      className="border-4 border-black bg-black text-white dark:border-white dark:bg-white dark:text-black"
      activeClassName="border-4 border-step-active-border-color dark:border-step-active-border-color"
      completedClassName="border-4 border-step-complete-border-color dark:border-step-complete-border-color"
      key={idx}
      onClick={() => setActiveStep(idx + 1)}
    >
      {idx + 2}
      <div className="absolute -bottom-[2rem] w-max text-center">
        <Typography variant="h6" className="text-black dark:text-gray-100">
          {group.display_name}
        </Typography>
      </div>
    </Step>
  ));

  let errorAlert = null;
  if (error.description) {
    errorAlert = (
      <ErrorAlert setError={setError}>{error.description}</ErrorAlert>
    );
  }

  return (
    <div className="relative w-full p-4">
      <Stepper
        activeStep={activeStep}
        isLastStep={(value) => setIsLastStep(value)}
        isFirstStep={(value) => setIsFirstStep(value)}
        lineClassName="bg-black dark:bg-white"
        activeLineClassName="!bg-step-complete-border-color"
      >
        <Step
          className="border-4 border-black bg-black text-white dark:border-white dark:bg-white dark:text-black"
          activeClassName="border-4 border-step-active-border-color dark:border-step-active-border-color"
          completedClassName="border-4 border-step-complete-border-color dark:border-step-complete-border-color"
          onClick={() => setActiveStep(0)}
        >
          1
          <div className="absolute -bottom-[2rem] w-max text-center">
            <Typography variant="h6" className="text-black dark:text-gray-100">
              Company Info
            </Typography>
          </div>
        </Step>

        {...steps}

        <Step
          className="border-4 border-black bg-black text-white dark:border-white dark:bg-white dark:text-black"
          activeClassName="border-4 border-step-active-border-color dark:border-step-active-border-color"
          completedClassName="border-4 border-step-complete-border-color dark:border-step-complete-border-color"
          onClick={() => setActiveStep(steps.length + 1)}
        >
          {steps.length + 2}
          <div className="absolute -bottom-[2rem] w-max text-center">
            <Typography variant="h6" className="text-black dark:text-gray-100">
              AWS IAM Role
            </Typography>
          </div>
        </Step>

        <Step
          className="border-4 border-black bg-black text-white dark:border-white dark:bg-white dark:text-black"
          activeClassName="border-4 border-step-active-border-color dark:border-step-active-border-color"
          completedClassName="border-4 border-step-complete-border-color dark:border-step-complete-border-color"
          onClick={() => setActiveStep(steps.length + 2)}
        >
          {steps.length + 3}
          <div className="absolute -bottom-[2rem] w-max text-center">
            <Typography variant="h6" className="text-black dark:text-gray-100">
              Install Status
            </Typography>
          </div>
        </Step>
      </Stepper>

      <form
        className="mt-10"
        onSubmit={formAction}
        onKeyUp={(e) => {
          e.preventDefault();
          if (e.key == "Enter") {
            handleNext();
          }
        }}
      >
        <CompanyContent
          name={install.name}
          open={activeStep === 0}
          onClick={() => setActiveStep(0)}
        />

        {...stepContent}

        <CloudAccountContent
          app={app}
          aws_account={install.aws_account}
          azure_account={install.aws_account}
          open={activeStep == steps.length + 1}
          onClick={() => setActiveStep(steps.length + 1)}
          searchParams={searchParams}
          regions={regions}
        />

        <InstallStatusContent
          open={activeStep === steps.length + 2}
          onClick={() => setActiveStep(steps.length + 2)}
          install={install}
          post_install_markdown={installer.metadata.post_install_markdown}
        />
      </form>

      <div className="mt-6">
        <NavArrows
          handlePrev={handlePrev}
          isFirstStep={isFirstStep}
          handleNext={handleNext}
          isLastStep={isLastStep}
        />
      </div>

      {errorAlert}
    </div>
  );
};

export default InstallStepper;
