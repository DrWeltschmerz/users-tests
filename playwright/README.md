# Playwright API Tests for users service

This test suite covers the same API paths as the Go integration tests, using Playwright for end-to-end HTTP testing.

## How to run

1. Install dependencies:
   npm install
2. Run tests:
   npx playwright test

## Structure
- `api.spec.ts`: Main Playwright test file covering user registration, login, admin, and role management endpoints.
- `playwright.config.ts`: Playwright configuration.

## Prerequisites
- The users service must be running and accessible (default: http://localhost:8080)
- Node.js and npm installed

---

Feel free to extend these tests for more coverage!
