const API = 'http://localhost:8080/api';

async function login(tcNo, password) {
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ tc_no: tcNo, password }),
  });
  return res.json();
}

async function logout() {
  await fetch(`${API}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  });
  window.location.href = 'index.html';
}

async function getMe() {
  const res = await fetch(`${API}/auth/me`, { credentials: 'include' });
  if (!res.ok) return null;
  return res.json();
}

// Login sayfasındaysak formu bağla
const loginForm = document.getElementById('login-form');
if (loginForm) {
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const tcNo    = document.getElementById('tc_no').value.trim();
    const password = document.getElementById('password').value;
    const errorBox = document.getElementById('error-box');
    const btn      = document.getElementById('login-btn');

    errorBox.classList.add('hidden');
    btn.disabled = true;
    btn.textContent = 'Giriş yapılıyor…';

    const data = await login(tcNo, password).catch(() => null);

    if (!data || !data.ok) {
      errorBox.textContent = data?.error || 'Bağlantı hatası.';
      errorBox.classList.remove('hidden');
      btn.disabled = false;
      btn.textContent = 'Giriş Yap';
      return;
    }

    // Role göre yönlendir
    if (data.member.role === 'moderator') {
      window.location.href = 'moderator.html';
    } else {
      window.location.href = 'voting.html';
    }
  });
}

// Çıkış butonu
const logoutBtn = document.getElementById('logout-btn');
if (logoutBtn) {
  logoutBtn.addEventListener('click', logout);
}