let currentVotingId = null;
let hasVoted = false;

function showWaiting() {
  document.getElementById('waiting-screen').classList.remove('hidden');
  document.getElementById('voting-screen').classList.add('hidden');
  document.getElementById('result-screen').classList.add('hidden');
}

function showVoting(title) {
  document.getElementById('waiting-screen').classList.add('hidden');
  document.getElementById('voting-screen').classList.remove('hidden');
  document.getElementById('result-screen').classList.add('hidden');
  document.getElementById('voting-title').textContent = title;
  document.getElementById('voted-msg').classList.add('hidden');
  document.getElementById('vote-buttons').classList.remove('hidden');
  hasVoted = false;
}

function showResult(result) {
  document.getElementById('waiting-screen').classList.add('hidden');
  document.getElementById('voting-screen').classList.add('hidden');
  document.getElementById('result-screen').classList.remove('hidden');

  const resultDisplay = document.getElementById('result-display');
  if (result === 'kabul_edildi') {
    resultDisplay.textContent = 'KABUL EDİLDİ';
    resultDisplay.className = 'result kabul';
  } else {
    resultDisplay.textContent = 'REDDEDİLDİ';
    resultDisplay.className = 'result red';
  }
}

async function castVote(choice) {
  if (hasVoted || !currentVotingId) return;

  document.getElementById('evet-btn').disabled  = true;
  document.getElementById('hayir-btn').disabled = true;

  const res = await fetch(`${API}/voting/${currentVotingId}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ choice }),
  });

  const data = await res.json();

  if (!res.ok) {
    alert(data.error || 'Oy kullanılamadı.');
    document.getElementById('evet-btn').disabled  = false;
    document.getElementById('hayir-btn').disabled = false;
    return;
  }

  hasVoted = true;
  document.getElementById('vote-buttons').classList.add('hidden');
  document.getElementById('voted-msg').classList.remove('hidden');
}

// WebSocket mesajları
ws.on('voting_started', (msg) => {
  currentVotingId = msg.voting_id;
  showVoting(msg.title);
});

ws.on('voting_closed', (msg) => {
  showResult(msg.result);
  currentVotingId = null;
});

// Başlat
ws.connect();
showWaiting();