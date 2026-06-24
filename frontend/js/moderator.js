let activeVotingId = null;

// Üye sayısını çek ve güncelle
async function refreshCount() {
  const res = await fetch(`${API}/attendance/count`, { credentials: 'include' });
  if (!res.ok) {
      console.error("HATA:", res.status, await res.text());
      return;
  }
  const data = await res.json();
  const count = data.present_count;

  document.getElementById('present-count').textContent = count;

  const badge   = document.getElementById('quorum-badge');
  const startBtn = document.getElementById('start-btn');
  const hint    = document.getElementById('start-hint');

  if (count >= 16) {
    badge.textContent = 'Yeter sayı sağlandı';
    badge.className   = 'badge badge-ready';
    startBtn.disabled = false;
    hint.classList.add('hidden');
  } else {
    badge.textContent = `Yeter sayı bekleniyor (${count}/16)`;
    badge.className   = 'badge badge-waiting';
    startBtn.disabled = true;
    hint.classList.remove('hidden');
  }
}

// Oylama başlat
document.getElementById('start-btn').addEventListener('click', async () => {
  const title = document.getElementById('voting-title').value.trim();
  if (!title) {
    alert('Lütfen gündem maddesini girin.');
    return;
  }

  const res = await fetch(`${API}/voting/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title }),
  });

  const data = await res.json();
  if (!res.ok) {
    alert(data.error || 'Oylama başlatılamadı.');
    return;
  }

  activeVotingId = data.id;
  document.getElementById('voting-title-display').textContent = data.title;
  document.getElementById('active-voting-card').classList.remove('hidden');
  document.getElementById('result-card').classList.add('hidden');
  document.getElementById('start-btn').disabled = true;
});

// Oylamayı kapat
document.getElementById('finalize-btn').addEventListener('click', async () => {
  if (!activeVotingId) return;
  if (!confirm('Oylamayı kapatmak istediğinize emin misiniz?')) return;

  const res = await fetch(`${API}/voting/${activeVotingId}/finalize`, {
    method: 'POST',
    credentials: 'include',
  });

  const data = await res.json();
  if (!res.ok) {
    alert(data.error || 'Oylama kapatılamadı.');
    return;
  }

  document.getElementById('active-voting-card').classList.add('hidden');

  const resultDisplay = document.getElementById('result-display');
  if (data.result === 'kabul_edildi') {
    resultDisplay.textContent = 'KABUL EDİLDİ';
    resultDisplay.className = 'result kabul';
  } else {
    resultDisplay.textContent = 'REDDEDİLDİ';
    resultDisplay.className = 'result red';
  }
  document.getElementById('result-card').classList.remove('hidden');
  activeVotingId = null;
});

// WebSocket mesajları
ws.on('vote_update', (msg) => {
  document.getElementById('yes-count').textContent = msg.yes;
  document.getElementById('no-count').textContent  = msg.no;
});

ws.on('voting_closed', (msg) => {
  document.getElementById('active-voting-card').classList.add('hidden');
  const resultDisplay = document.getElementById('result-display');
  if (msg.result === 'kabul_edildi') {
    resultDisplay.textContent = 'KABUL EDİLDİ';
    resultDisplay.className = 'result kabul';
  } else {
    resultDisplay.textContent = 'REDDEDİLDİ';
    resultDisplay.className = 'result red';
  }
  document.getElementById('result-card').classList.remove('hidden');
});

// Başlat
ws.connect();
refreshCount();
setInterval(refreshCount, 5000);