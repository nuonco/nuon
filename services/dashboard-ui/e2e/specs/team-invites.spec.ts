import { test, expect } from "../fixtures";

test.describe("Team invites", () => {
  test.setTimeout(60000);

  test("invite a user, resend, then revoke", async ({ page, orgId }) => {
    const email = `e2e-invite-${Date.now()}@example.com`;

    await page.goto(`/${orgId}/team`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Team \|/, { timeout: 15000 });

    // --- Invite ---
    await page.getByRole("button", { name: "Invite user" }).first().click();
    const inviteDialog = page.getByRole("dialog");
    await expect(inviteDialog.getByText("Invite team member")).toBeVisible();
    await inviteDialog.getByPlaceholder("user@email.com").fill(email);
    await inviteDialog.getByRole("button", { name: "Invite user" }).click();
    await expect(page.getByText("Invitation sent")).toBeVisible({
      timeout: 10000,
    });

    const row = page
      .locator("div.flex.items-center.gap-4")
      .filter({ hasText: email });
    await expect(row).toBeVisible({ timeout: 10000 });

    // --- Resend ---
    await row.getByRole("button", { name: "Resend" }).click();
    const resendDialog = page.getByRole("dialog");
    await expect(
      resendDialog.getByRole("button", { name: "Resend invite" })
    ).toBeVisible();
    await resendDialog.getByRole("button", { name: "Resend invite" }).click();
    await expect(page.getByText("Invite resent")).toBeVisible({ timeout: 10000 });

    // --- Revoke ---
    await row.getByRole("button", { name: "Revoke" }).click();
    const revokeDialog = page.getByRole("dialog");
    await expect(revokeDialog.getByText("Revoke invite?")).toBeVisible();
    await revokeDialog.getByRole("button", { name: "Revoke invite" }).click();
    await expect(page.getByText("Invite revoked")).toBeVisible({
      timeout: 10000,
    });
    await expect(row).not.toBeVisible();
  });
});
