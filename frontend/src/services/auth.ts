import React from "react";
import { notify } from "@/lib/notifier";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;
const GOOGLE_LOGIN_API = import.meta.env.VITE_GOOGLE_LOGIN;

/**
 * Handles email login form submission.
 *
 * Expected form fields:
 * - input-email
 * - input-password
 */
async function emailLogin(e: React.FormEvent<HTMLFormElement>) {
  e.preventDefault();

  const formData = new FormData(e.currentTarget);

  const email = formData.get("input-email");
  const password = formData.get("input-password");

  if (!password || !email) {
    notify.error("Missing email or password");
    return;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: String(email),
        password: String(password),
      }),
    });

    let data = null;

    const text = await res.text();

    if (text) {
      data = JSON.parse(text);
    }

    if (!res.ok) {
      notify.error("Login failed: " + data.error);
      return;
    }

    const token = data.token;
    localStorage.setItem("token", token);

    notify.news("Login success: " + data);
    window.location.reload();
  } catch {
    notify.error("Login failed");
  }
}

/**
 * Handles email register form submission.
 *
 * Expected form fields:
 * - input-username
 * - input-email
 * - input-password
 */
async function emailRegister(e: React.FormEvent<HTMLFormElement>) {
  e.preventDefault();

  const formData = new FormData(e.currentTarget);

  const username = formData.get("input-username");
  const email = formData.get("input-email");
  const password = formData.get("input-password");

  if (!username || !email || !password) {
    notify.error("Missing values");
    return;
  }

  try {
    const res = await fetch(`${BACKEND_URL}/api/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: String(email),
        username: String(username),
        password: String(password),
      }),
    });

    let data = null;

    const text = await res.text();

    if (text) {
      data = JSON.parse(text);
    }

    if (!res.ok) {
      notify.error("Register failed: " + data.error);
      return;
    }

    notify.news("Register success");
  } catch {
    notify.error("Register failed");
  }
}

function googleLogin() {
  window.location.href = GOOGLE_LOGIN_API;
}

export { emailLogin, emailRegister, googleLogin };
