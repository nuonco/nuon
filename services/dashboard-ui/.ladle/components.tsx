import "../client/styles.css";
import { GlobalProvider } from "@ladle/react";
import { MemoryRouter } from "react-router";

export const Provider: GlobalProvider = ({ children }) => {
  return <MemoryRouter>{children}</MemoryRouter>;
};
