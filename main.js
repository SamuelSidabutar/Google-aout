// ============================================================
// Google Auth Sign-In dengan CrootJS
// Frontend logic menggunakan JSCroot modules
// ============================================================

// Import modul CrootJS
import { postJSON } from "https://cdn.jsdelivr.net/gh/jscroot/lib@0.2.8/api.js";
import { setCookieWithExpireHour, getCookie, deleteCookie } from "https://cdn.jsdelivr.net/gh/jscroot/lib@0.2.8/cookie.js";

// ============================================================
// Konfigurasi
// ============================================================
const BACKEND_URL = "http://localhost:3000/api/auth/google";

// ============================================================
// Google Sign-In Callback
// Fungsi ini dipanggil otomatis oleh Google Identity Services
// setelah user berhasil login
// ============================================================
function handleCredentialResponse(response) {
    console.log("📥 Menerima credential dari Google...");

    // Tampilkan loading state
    showLoading(true);

    // Ambil JWT ID Token dari response Google
    const idToken = response.credential;

    // Kirim token ke backend Golang untuk verifikasi
    // postJSON signature: postJSON(url, datajson, responseFunction)
    postJSON(
        BACKEND_URL,
        { token: idToken },
        function (result) {
            handleAuthResponse(result);
        }
    );
}

// ============================================================
// Handle Response dari Backend
// ============================================================
function handleAuthResponse(result) {
    showLoading(false);

    console.log("📨 Response dari backend:", result);

    if (result.status === 200 && result.data.status === "success") {
        // Login berhasil!
        const userData = result.data.data;
        const token = result.data.token;

        console.log("✅ Login berhasil:", userData.name);

        // Simpan token di cookie selama 2 jam menggunakan CrootJS
        setCookieWithExpireHour("login", token, 2);

        // Simpan data user di cookie untuk ditampilkan di dashboard
        setCookieWithExpireHour("user_name", userData.name, 2);
        setCookieWithExpireHour("user_email", userData.email, 2);
        setCookieWithExpireHour("user_picture", userData.picture, 2);

        // Tampilkan notifikasi sukses
        showNotification("success", `Selamat datang, ${userData.name}! 🎉`);

        // Redirect ke dashboard setelah 1.5 detik
        setTimeout(() => {
            window.location.href = "dashboard.html";
        }, 1500);
    } else {
        // Login gagal
        const message = result.data?.message || "Login gagal. Silakan coba lagi.";
        console.error("❌ Login gagal:", message);
        showNotification("error", message);
    }
}

// ============================================================
// UI Helper Functions
// ============================================================

// Tampilkan/sembunyikan loading spinner
function showLoading(show) {
    const loader = document.getElementById("loading-overlay");
    if (loader) {
        loader.style.display = show ? "flex" : "none";
    }
}

// Tampilkan notifikasi toast
function showNotification(type, message) {
    const container = document.getElementById("notification-container");
    if (!container) return;

    const notification = document.createElement("div");
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <span class="notification-icon">${type === "success" ? "✅" : "❌"}</span>
        <span class="notification-text">${message}</span>
    `;

    container.appendChild(notification);

    // Animasi masuk
    requestAnimationFrame(() => {
        notification.classList.add("show");
    });

    // Hapus setelah 4 detik
    setTimeout(() => {
        notification.classList.remove("show");
        setTimeout(() => notification.remove(), 300);
    }, 4000);
}

// ============================================================
// Cek apakah user sudah login (untuk redirect otomatis)
// ============================================================
function checkExistingLogin() {
    const token = getCookie("login");
    if (token && token !== "") {
        console.log("🔄 User sudah login, redirect ke dashboard...");
        window.location.href = "dashboard.html";
    }
}

// ============================================================
// Logout Function (digunakan di dashboard)
// ============================================================
function logout() {
    deleteCookie("login");
    deleteCookie("user_name");
    deleteCookie("user_email");
    deleteCookie("user_picture");

    console.log("👋 Logout berhasil");
    window.location.href = "index.html";
}

// ============================================================
// Expose functions ke global scope
// ============================================================
window.handleCredentialResponse = handleCredentialResponse;
window.logout = logout;

// Cek login saat halaman dimuat (hanya di halaman login)
if (window.location.pathname.endsWith("index.html") || window.location.pathname.endsWith("/")) {
    checkExistingLogin();
}
