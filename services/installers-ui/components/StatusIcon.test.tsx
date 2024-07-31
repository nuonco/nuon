import { expect, test } from "vitest";
import { render } from "@testing-library/react";
import { StatusIcon } from "./StatusIcon";

test("Renders green dot for active status", () => {
  const { container } = render(<StatusIcon status="active" />);
  console.log(container.firstChild.className);
  expect(container.firstChild.className.includes("green-600")).toBeTruthy();
});

test("Renders red dot for error status", () => {
  const { container } = render(<StatusIcon status="error" />);
  expect(container.firstChild.className.includes("red-500")).toBeTruthy();
});

test("Renders yellow dot for any other string", () => {
  const { container } = render(<StatusIcon status="pending" />);
  expect(container.firstChild.className.includes("yellow-500")).toBeTruthy();
});

test("Renders gray dot for empty string", () => {
  const { container } = render(<StatusIcon status="" />);
  expect(container.firstChild.className.includes("slate-300")).toBeTruthy();
});
