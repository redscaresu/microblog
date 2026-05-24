const { defineConfig } = require("@playwright/test");

module.exports = defineConfig({
  testDir: "./tests",
  timeout: 60000,
  expect: {
    timeout: 10000,
  },
  use: {
    baseURL: "http://127.0.0.1:18080",
    httpCredentials: {
      username: "foo",
      password: "foo",
    },
    trace: "on-first-retry",
  },
  webServer: {
    command: "go run ./cmd/e2e-server",
    url: "http://127.0.0.1:18080",
    timeout: 120000,
    reuseExistingServer: false,
  },
});
