import { expect, test } from "@playwright/test";

const WITH_GROUPS =
  "/?story=branches--deploymentplaneditor--with-groups&mode=preview";
const NO_GROUPS =
  "/?story=branches--deploymentplaneditor--no-groups&mode=preview";

const openEditor = async (page, story: string) => {
  await page.goto(story, { waitUntil: "domcontentloaded" });
  await page.getByRole("button", { name: "Open deployment plan" }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.getByRole("dialog")).toBeFocused();
};

test.describe("DeploymentPlanEditor behavior", () => {
  test("adding a group focuses its name field", async ({ page }) => {
    await openEditor(page, WITH_GROUPS);
    const dialog = page.getByRole("dialog");

    await expect(dialog.getByLabel("Group 1 name")).not.toBeFocused();

    await dialog.getByRole("button", { name: "Add group" }).click();
    const newName = dialog.getByLabel("Group 2 name");
    await expect(newName).toBeFocused();

    await page.keyboard.type("canaries");
    await expect(newName).toHaveValue("canaries");
    await expect(dialog.getByLabel("Group 1 name")).toHaveValue("Production");
  });

  test("the first group added to an empty plan is focused too", async ({
    page,
  }) => {
    await openEditor(page, NO_GROUPS);
    const dialog = page.getByRole("dialog");

    await dialog.getByRole("button", { name: "Add group" }).click();
    await expect(dialog.getByLabel("Group 1 name")).toBeFocused();
  });
});
