import React from "react";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL
const GOOGLE_LOGIN_API = import.meta.env.VITE_GOOGLE_LOGIN

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
        processError("Missing email or password")
        return
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

        if (!res.ok) {
            processError("Login failed");
            return;
        }

        const data = await res.json();
        console.log("Login success:", data);
    } catch {
        processError("Network error");
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
function emailRegister(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    const formData = new FormData(e.currentTarget);

    const username = formData.get("input-username");
    const email = formData.get("input-email");
    const password = formData.get("input-password");
} 

function googleLogin() {
    window.location.href = GOOGLE_LOGIN_API;
}

function processError(msg: string) {
    console.error(msg) //TODO: Implement UI pop up
}

export {emailLogin, emailRegister, googleLogin}