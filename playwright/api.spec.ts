import { expect, test } from "@playwright/test";

const testUser = {
  email: "testuser@example.com",
  username: "testuser",
  password: "testpass",
};
const adminUser = {
  email: "admin@example.com",
  username: "admin",
  password: "adminpass",
};

let userToken = "";
let adminToken = "";

// Helper to register a user
async function registerUser(api, user) {
  return api.post("/register", { data: user });
}

// Helper to login and get JWT
async function loginUser(api, user) {
  // The Go API expects email and password for login
  const res = await api.post("/login", {
    data: { email: user.email, password: user.password },
  });
  return res.ok() ? res.json().then((j) => j.token) : "";
}

test.describe("API Integration", () => {
  test("Register user", async ({ request }) => {
    const res = await registerUser(request, testUser);
    if (!res.ok()) {
      console.log("Register user failed:", res.status(), await res.text());
    }
    expect(res.ok()).toBeTruthy();
  });

  test("Login user", async ({ request }) => {
    userToken = await loginUser(request, testUser);
    if (!userToken) {
      console.log("Login user failed: token is empty");
    }
    expect(userToken).not.toBe("");
  });

  // No need to register admin, it is seeded in the test server

  test("Login admin", async ({ request }) => {
    adminToken = await loginUser(request, adminUser);
    if (!adminToken) {
      console.log("Login admin failed: token is empty");
    }
    expect(adminToken).not.toBe("");
  });

  test("User cannot access admin endpoint", async ({ request }) => {
    const res = await request.get("/users", {
      headers: { Authorization: `Bearer ${userToken}` },
    });
    if (res.status() !== 403) {
      console.log(
        "User cannot access admin endpoint:",
        res.status(),
        await res.text()
      );
    }
    expect(res.status()).toBe(403);
  });

  test("Admin can access admin endpoint", async ({ request }) => {
    const res = await request.get("/users", {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    if (!res.ok()) {
      console.log(
        "Admin can access admin endpoint failed:",
        res.status(),
        await res.text()
      );
    }
    expect(res.ok()).toBeTruthy();
    expect(Array.isArray(await res.json())).toBeTruthy();
  });

  test("User can get own profile", async ({ request }) => {
    const res = await request.get("/user/profile", {
      headers: { Authorization: `Bearer ${userToken}` },
    });
    if (!res.ok()) {
      console.log(
        "User can get own profile failed:",
        res.status(),
        await res.text()
      );
    }
    expect(res.ok()).toBeTruthy();
    const profile = await res.json();
    expect(profile.Username).toBe(testUser.username);
  });

  // Add more tests for role management, etc. as needed
});
