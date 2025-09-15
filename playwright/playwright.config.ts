import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./",
  timeout: 30000,
  retries: 0,
  reporter: [["html", { outputFolder: "playwright-report", open: "never" }]],
  use: {
    baseURL: "http://localhost:8080",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "API tests",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
