const { test, expect, request } = require("@playwright/test");

function slugify(title) {
  return title.replaceAll(" ", "-").toLowerCase().replace(/[^a-z0-9\s]+/g, "");
}

test("create, render, update, and delete a blog post", async ({ page, baseURL }) => {
  const title = `Playwright happy path ${Date.now()}`;
  const content = "Initial content from Playwright.";
  const updatedTitle = `${title} updated`;
  const updatedContent = "Updated content from Playwright.";
  const slug = slugify(title);

  await page.goto("/admin/post/new");
  await page.getByLabel("Title:").fill(title);
  await page.getByLabel("Content:").fill(content);
  const createResponsePromise = page.waitForResponse((response) =>
    response.url().endsWith("/api/post/new") && response.request().method() === "POST"
  );
  await page.getByRole("button", { name: "Submit" }).click();

  const createResponse = await createResponsePromise;
  expect(createResponse.ok()).toBeTruthy();
  const createBody = await createResponse.text();
  const createdPost = JSON.parse(createBody.slice(0, createBody.indexOf("}") + 1));
  expect(createdPost.Name).toBe(slug);

  await page.goto("/");
  await expect(page.getByRole("link", { name: title })).toBeVisible();
  await expect(page.locator("#blog-container")).toContainText(content);

  await page.goto(`/post/${createdPost.Name}`);
  await expect(page.locator("h1")).toContainText(title);
  await expect(page.locator("#blog-post-container")).toContainText(content);

  await page.goto(`/admin/post/edit/${createdPost.Name}`);
  await page.getByLabel("Title:").fill(updatedTitle);
  await page.evaluate((value) => {
    const textarea = document.getElementById("content");
    textarea.value = value;
    if (window.simplemde) {
      window.simplemde.value(value);
    }
  }, updatedContent);
  await page.getByRole("button", { name: "Submit" }).click();

  await expect(page.locator("body")).toContainText("Post updated successfully!");

  await page.goto("/");
  await expect(page.getByRole("link", { name: updatedTitle })).toBeVisible();
  await expect(page.locator("#blog-container")).toContainText(updatedContent);

  await page.goto(`/post/${createdPost.Name}`);
  await expect(page.locator("h1")).toContainText(updatedTitle);
  await expect(page.locator("#blog-post-container")).toContainText(updatedContent);

  const apiContext = await request.newContext({
    baseURL,
    httpCredentials: {
      username: "foo",
      password: "foo",
    },
  });

  const deleteResponse = await apiContext.post(`/api/post/delete/${createdPost.ID}`);
  expect(deleteResponse.ok()).toBeTruthy();
  await apiContext.dispose();

  await page.goto("/");
  await expect(page.getByRole("link", { name: updatedTitle })).toHaveCount(0);
});
