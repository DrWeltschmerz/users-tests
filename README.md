# users-tests

Integration tests for the DrWeltschmerz users modules (users-core, users-adapter-gin, users-adapter-gorm, jwt-auth).

## What is tested?
- User registration, login, profile, and password change
- Role-based access control (admin/user)
- Admin-only endpoints and permissions
- Negative cases (invalid input, duplicate registration, unauthorized access)

## How to run

1. Clone this repo:
   ```sh
   git clone git@github.com:DrWeltschmerz/users-tests.git
   cd users-tests
   ```
2. Ensure you have Go installed (>=1.20).
3. Run the tests:
   ```sh
   go test -v ./...
   ```

### Playwright API tests

From `playwright/`:

```sh
cd playwright
npm install
npx playwright test
```

By default tests expect the users service on http://localhost:8080. For local runs, you can start the bundled test server:

```sh
cd server
go run main.go
```

## CI
These tests are intended to be run in CI on every push and on new tags of any related users module. See `.github/workflows/ci.yml` for the GitHub Actions setup.

CI uploads a Playwright HTML report as an artifact (folder `playwright-report/`). Download it from the Actions run to inspect failures.

## Requirements
- All related modules must be accessible (either via replace directives or as tagged versions on GitHub).
- No external database needed; tests use in-memory SQLite.

Related modules:
- [users-core](https://github.com/DrWeltschmerz/users-core)
- [users-adapter-gin](https://github.com/DrWeltschmerz/users-adapter-gin)
- [users-adapter-gorm](https://github.com/DrWeltschmerz/users-adapter-gorm)
- [jwt-auth](https://github.com/DrWeltschmerz/jwt-auth)

---
Maintained by DrWeltschmerz
