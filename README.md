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

## CI
These tests are intended to be run in CI on every push and on new tags of any related users module. See `.github/workflows/ci.yml` for the GitHub Actions setup.

## Requirements
- All related modules must be accessible (either via replace directives or as tagged versions on GitHub).
- No external database needed; tests use in-memory SQLite.

---
Maintained by DrWeltschmerz
