const esikler = {
  salt_cogunluk: 17,
  iki_uc: 22,
  uc_dort: 24,
  oybirligi: 32,
};

let activeVotingId = null;

// ── Sayfa yönlendirme ──
document.querySelectorAll('.mod-nav-item').forEach(item => {
  item.addEventListener('click', () => {
    document.querySelectorAll('.mod-nav-item').forEach(i => i.classList.remove('active'));
    document.querySelectorAll('.mod-page').forEach(p => p.classList.add('hidden'));
    item.classList.add('active');
    document.getElementById('page-' + item.dataset.page).classList.remove('hidden');
    document.getElementById('sidebar').classList.remove('open');
    if (item.dataset.page === 'panel')   loadPanel();
    if (item.dataset.page === 'yoklama') loadYoklama();
    if (item.dataset.page === 'uyeler')  loadUyeler();
    if (item.dataset.page === 'raporlar') loadRaporlar();
    if (item.dataset.page === 'huzur-hakki') loadHuzur();
    if (item.dataset.page === 'oturumlar') loadOturumlar();
  });
});

document.getElementById('menu-toggle').addEventListener('click', () => {
  document.getElementById('sidebar').classList.toggle('open');
});

// ── Panel ──
async function loadPanel() {
  const [countRes, activeRes, recentRes] = await Promise.all([
    fetch(`${API}/attendance/count`, { credentials: 'include' }),
    fetch(`${API}/voting/active`,    { credentials: 'include' }),
    fetch(`${API}/voting/recent`,    { credentials: 'include' }),
  ]);

  if (countRes.ok) {
    const data  = await countRes.json();
    const count = data.present_count;
    document.getElementById('stat-present').textContent = count;
    document.getElementById('stat-quorum').textContent  = count >= 16 ? 'Sağlandı ✓' : 'Bekleniyor';
  }

  if (activeRes.ok) {
    const data = await activeRes.json();
    if (data.active) {
      activeVotingId = data.id;
      document.getElementById('stat-voting').textContent          = 'Aktif ✓';
      document.getElementById('voting-title-display').textContent = data.title;
      document.getElementById('active-oylama-tipi').textContent   = data.oylama_tipi === 'gizli' ? 'Gizli' : 'Açık';
      document.getElementById('active-esik').textContent          = `${data.esik_sayi} oy`;
      document.getElementById('yes-count').textContent            = data.yes;
      document.getElementById('no-count').textContent             = data.no;
      document.getElementById('active-voting-card').classList.remove('hidden');
      document.getElementById('start-btn').disabled = true;
    } else {
      activeVotingId = null;
      document.getElementById('stat-voting').textContent = 'Yok';
    }
  }

  if (recentRes.ok) {
    const votings = await recentRes.json();
    const el = document.getElementById('recent-votings');
    if (votings.length === 0) {
      el.innerHTML = '<p class="hint">Henüz oylama yapılmadı.</p>';
    } else {
      el.innerHTML = `
        <table class="mod-table">
          <thead>
            <tr><th>Başlık</th><th>Tip</th><th>Eşik</th><th>Durum</th><th>Sonuç</th></tr>
          </thead>
          <tbody>
            ${votings.map(v => `
              <tr>
                <td>${v.title}</td>
                <td>${v.oylama_tipi === 'gizli' ? 'Gizli' : 'Açık'}</td>
                <td>${v.esik_sayi}</td>
                <td>${v.status === 'active' ? '🟢 Aktif' : '⚫ Kapandı'}</td>
                <td>${v.result === 'kabul_edildi' ? '✅ Kabul' : v.result === 'reddedildi' ? '❌ Red' : '—'}</td>
              </tr>
            `).join('')}
          </tbody>
        </table>`;
    }
  }
}

// ── Yoklama ──
async function loadYoklama() {
  const res = await fetch(`${API}/attendance/count`, { credentials: 'include' });
  if (!res.ok) return;
  const data  = await res.json();
  const count = data.present_count;

  document.getElementById('present-count').textContent = count;
  const badge = document.getElementById('quorum-badge');
  if (count >= 16) {
    badge.textContent = `Yeter sayı sağlandı (${count}/32)`;
    badge.className   = 'badge badge-ready';
  } else {
    badge.textContent = `Yeter sayı bekleniyor (${count}/16)`;
    badge.className   = 'badge badge-waiting';
  }

  const listEl = document.getElementById('salon-uye-listesi');
  const uyeRes = await fetch(`${API}/attendance/present`, { credentials: 'include' }).catch(() => null);
  if (uyeRes && uyeRes.ok) {
    const uyeler = await uyeRes.json();
    if (uyeler.length === 0) {
      listEl.innerHTML = '<p class="hint">Salonda üye yok.</p>';
    } else {
      listEl.innerHTML = `
        <table class="mod-table">
          <thead><tr><th>ID</th><th>Ad Soyad</th><th>Parti</th></tr></thead>
          <tbody>
            ${uyeler.map(u => `
              <tr>
                <td>${u.id}</td>
                <td>${u.name} ${u.soyisim || ''}</td>
                <td>${u.parti || '—'}</td>
              </tr>
            `).join('')}
          </tbody>
        </table>`;
    }
  } else {
    listEl.innerHTML = '<p class="hint">Üye listesi alınamadı.</p>';
  }
}

// ── Üyeler ──
async function loadUyeler() {
  const listEl = document.getElementById('uye-listesi');
  const res    = await fetch(`${API}/members`, { credentials: 'include' }).catch(() => null);

  if (!res || !res.ok) {
    listEl.innerHTML = '<p class="hint">Üye listesi alınamadı.</p>';
    return;
  }

  const uyeler = await res.json();
  listEl.innerHTML = `
    <table class="mod-table">
      <thead>
        <tr><th>ID</th><th>Ad</th><th>Soyad</th><th>TC No</th><th>Parti</th><th>Rol</th></tr>
      </thead>
      <tbody>
        ${uyeler.map(u => `
          <tr>
            <td>${u.id}</td>
            <td>${u.name}</td>
            <td>${u.soyisim || '—'}</td>
            <td>${u.tc_no}</td>
            <td>${u.parti || '—'}</td>
            <td>${u.role}</td>
          </tr>
        `).join('')}
      </tbody>
    </table>`;
}

document.getElementById('uye-ara').addEventListener('input', (e) => {
  const q = e.target.value.toLowerCase();
  document.querySelectorAll('#uye-listesi tbody tr').forEach(row => {
    row.style.display = row.textContent.toLowerCase().includes(q) ? '' : 'none';
  });
});

// ── Oylama ──
document.getElementById('esik-tipi').addEventListener('change', (e) => {
  document.getElementById('esik-sayi').textContent = esikler[e.target.value];
});

async function refreshOylama() {
  const res = await fetch(`${API}/attendance/count`, { credentials: 'include' });
  if (!res.ok) return;
  const data  = await res.json();
  const count = data.present_count;
  const startBtn = document.getElementById('start-btn');
  const hint     = document.getElementById('start-hint');
  if (count >= 16) {
    startBtn.disabled = false;
    hint.classList.add('hidden');
  } else {
    startBtn.disabled = true;
    hint.classList.remove('hidden');
  }
}

document.getElementById('start-btn').addEventListener('click', async () => {
  const title      = document.getElementById('voting-title').value.trim();
  const oylamaTipi = document.querySelector('input[name="oylama_tipi"]:checked').value;
  const esikTipi   = document.getElementById('esik-tipi').value;

  if (!title) { alert('Lütfen gündem maddesini girin.'); return; }

  const res = await fetch(`${API}/voting/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title, oylama_tipi: oylamaTipi, esik_tipi: esikTipi }),
  });

  const data = await res.json();
  if (!res.ok) { alert(data.error || 'Oylama başlatılamadı.'); return; }

  activeVotingId = data.id;
  document.getElementById('voting-title-display').textContent = data.title;
  document.getElementById('active-oylama-tipi').textContent   = data.oylama_tipi === 'gizli' ? 'Gizli' : 'Açık';
  document.getElementById('active-esik').textContent          = `${esikler[data.esik_tipi]} oy`;
  document.getElementById('yes-count').textContent            = '0';
  document.getElementById('no-count').textContent             = '0';
  document.getElementById('active-voting-card').classList.remove('hidden');
  document.getElementById('result-card').classList.add('hidden');
  document.getElementById('start-btn').disabled = true;
  document.getElementById('voting-title').value = '';
});

document.getElementById('finalize-btn').addEventListener('click', async () => {
  if (!activeVotingId) return;
  if (!confirm('Oylamayı kapatmak istediğinize emin misiniz?')) return;

  const res  = await fetch(`${API}/voting/${activeVotingId}/finalize`, {
    method: 'POST',
    credentials: 'include',
  });
  const data = await res.json();
  if (!res.ok) { alert(data.error || 'Oylama kapatılamadı.'); return; }
  showResult(data);
});

function showResult(data) {
  document.getElementById('active-voting-card').classList.add('hidden');
  const rd = document.getElementById('result-display');
  if (data.result === 'kabul_edildi') {
    rd.textContent = 'KABUL EDİLDİ';
    rd.className   = 'result kabul';
  } else {
    rd.textContent = 'REDDEDİLDİ';
    rd.className   = 'result red';
  }
  if (data.yes_count !== undefined) {
    document.getElementById('result-detail').textContent =
      `${data.yes_count} EVET · Eşik: ${data.esik_sayi}`;
  }
  document.getElementById('result-card').classList.remove('hidden');
  activeVotingId = null;
}

// ── WebSocket ──
ws.on('vote_update', (msg) => {
  document.getElementById('yes-count').textContent = msg.yes;
  document.getElementById('no-count').textContent  = msg.no;
});

ws.on('voting_closed',  (msg) => showResult(msg));
ws.on('voting_started', ()    => document.getElementById('stat-voting').textContent = 'Aktif ✓');

// ── Başlat ──
ws.connect();
loadPanel();
refreshOylama();

// ── Raporlar ──
async function loadRaporlar() {
  const res = await fetch(`${API}/reports/votings`, { credentials: 'include' });
  const el  = document.getElementById('rapor-listesi');

  if (!res.ok) { el.innerHTML = '<p class="hint">Raporlar alınamadı.</p>'; return; }

  const list = await res.json();
  if (list.length === 0) {
    el.innerHTML = '<p class="hint">Henüz oylama yapılmadı.</p>';
    return;
  }

  el.innerHTML = `
    <table class="mod-table">
      <thead>
        <tr>
          <th>#</th><th>Başlık</th><th>Tip</th><th>Eşik</th>
          <th>EVET</th><th>HAYIR</th><th>Sonuç</th><th>Detay</th>
        </tr>
      </thead>
      <tbody>
        ${list.map(v => `
          <tr>
            <td>${v.id}</td>
            <td>${v.title}</td>
            <td>${v.oylama_tipi === 'gizli' ? 'Gizli' : 'Açık'}</td>
            <td>${v.esik_sayi}</td>
            <td style="color:var(--success);font-weight:600">${v.yes_count}</td>
            <td style="color:var(--danger);font-weight:600">${v.no_count}</td>
            <td>${v.result === 'kabul_edildi' ? '✅ Kabul' : v.result === 'reddedildi' ? '❌ Red' : v.status === 'active' ? '🟢 Aktif' : '—'}</td>
            <td>
              ${v.oylama_tipi === 'acik'
                ? `<button class="btn-ghost" style="padding:0.25rem 0.5rem;font-size:0.8rem" onclick="openDetay(${v.id}, '${v.title}')">Görüntüle</button>`
                : '<span class="hint">Gizli</span>'
              }
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>`;
}

async function openDetay(votingId, title) {
  document.getElementById('detay-baslik').textContent = title;
  document.getElementById('detay-icerik').innerHTML   = 'Yükleniyor…';
  document.getElementById('detay-modal').classList.remove('hidden');

  const res = await fetch(`${API}/reports/votings/${votingId}/detail`, { credentials: 'include' });
  const el  = document.getElementById('detay-icerik');

  if (!res.ok) { el.innerHTML = '<p class="hint">Detay alınamadı.</p>'; return; }

  const list = await res.json();
  if (list.length === 0) { el.innerHTML = '<p class="hint">Oy kullanılmamış.</p>'; return; }

  el.innerHTML = `
    <table class="mod-table">
      <thead>
        <tr><th>Ad Soyad</th><th>Parti</th><th>Oy</th></tr>
      </thead>
      <tbody>
        ${list.map(d => `
          <tr>
            <td>${d.name} ${d.soyisim}</td>
            <td>${d.parti || '—'}</td>
            <td style="font-weight:700;color:${d.choice === 'evet' ? 'var(--success)' : 'var(--danger)'}">
              ${d.choice === 'evet' ? 'EVET' : 'HAYIR'}
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>`;
}

function closeDetay() {
  document.getElementById('detay-modal').classList.add('hidden');
}

// ── Huzur Hakkı ──
let huzurRoles = [];
let selectedMemberForRole = null;

async function loadHuzur() {
  const [listRes, rolesRes] = await Promise.all([
    fetch(`${API}/huzur/list`, { credentials: 'include' }),
    fetch(`${API}/huzur/roles`, { credentials: 'include' }),
  ]);

  if (rolesRes.ok) {
    huzurRoles = await rolesRes.json();
    const select = document.getElementById('rol-select');
    select.innerHTML = huzurRoles.map(r =>
      `<option value="${r.id}">${r.role_name} (×${r.katsayi})</option>`
    ).join('');
  }

  if (!listRes.ok) return;
  const data = await listRes.json();

  // Taban tutarı göster
  document.getElementById('taban-tutar').value = data.taban_tutar || '';

  // Toplam hesapla
  const toplam = data.list.filter(r => r.katildi).reduce((s, r) => s + r.hakkedis, 0);
  document.getElementById('huzur-toplam').textContent =
    `Toplam: ${toplam.toLocaleString('tr-TR', { style: 'currency', currency: 'TRY' })}`;

  const el = document.getElementById('huzur-listesi');
  if (data.list.length === 0) {
    el.innerHTML = '<p class="hint">Üye bulunamadı.</p>';
    return;
  }

  el.innerHTML = `
    <table class="mod-table">
      <thead>
        <tr><th>Ad Soyad</th><th>Rol</th><th>Katsayı</th><th>Katıldı</th><th>Hak Ediş</th><th>İşlem</th></tr>
      </thead>
      <tbody>
        ${data.list.map(r => `
          <tr>
            <td>${r.name} ${r.soyisim}</td>
            <td>${r.role_name}</td>
            <td>×${r.katsayi}</td>
            <td>${r.katildi ? '✅' : '❌'}</td>
            <td style="font-weight:600;color:${r.katildi ? 'var(--success)' : 'var(--text-muted)'}">
              ${r.katildi
                ? r.hakkedis.toLocaleString('tr-TR', { style: 'currency', currency: 'TRY' })
                : '—'}
            </td>
            <td>
              <button class="btn-ghost" style="font-size:0.8rem;padding:0.2rem 0.5rem"
                onclick="openRolModal(${r.member_id}, '${r.name} ${r.soyisim}')">
                Rol Ata
              </button>
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>`;
}

async function saveHuzurSettings() {
  const tutar = parseFloat(document.getElementById('taban-tutar').value);
  if (!tutar || tutar <= 0) { alert('Geçerli bir tutar girin.'); return; }

  const res = await fetch(`${API}/huzur/settings`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ taban_tutar: tutar }),
  });

  if (res.ok) { loadHuzur(); } else { alert('Kayıt başarısız.'); }
}

function openRolModal(memberId, name) {
  selectedMemberForRole = memberId;
  document.getElementById('rol-modal-uye').textContent = name;
  document.getElementById('rol-modal').classList.remove('hidden');
}

function closeRolModal() {
  document.getElementById('rol-modal').classList.add('hidden');
  selectedMemberForRole = null;
}

async function saveRol() {
  if (!selectedMemberForRole) return;
  const roleId = parseInt(document.getElementById('rol-select').value);

  const res = await fetch(`${API}/huzur/member-role`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ member_id: selectedMemberForRole, role_id: roleId }),
  });

  if (res.ok) { closeRolModal(); loadHuzur(); } else { alert('Rol atanamadı.'); }
}

// ── Oturumlar ──
async function loadOturumlar() {
  const res = await fetch(`${API}/meetings`, { credentials: 'include' });
  const el  = document.getElementById('oturum-listesi');

  if (!res.ok) { el.innerHTML = '<p class="hint">Oturumlar alınamadı.</p>'; return; }

  const list = await res.json();
  if (list.length === 0) {
    el.innerHTML = '<p class="hint">Henüz oturum planlanmadı.</p>';
    return;
  }

  const statusLabel = {
    planned: '<span style="color:var(--warning);font-weight:600">📅 Planlandı</span>',
    active:  '<span style="color:var(--success);font-weight:600">🟢 Aktif</span>',
    ended:   '<span style="color:var(--text-muted)">⚫ Kapandı</span>',
  };

  el.innerHTML = `
    <table class="mod-table">
      <thead>
        <tr><th>No</th><th>Başlık</th><th>Tarih</th><th>Durum</th><th>İşlem</th></tr>
      </thead>
      <tbody>
        ${list.map(m => `
          <tr>
            <td>${m.meeting_no || '—'}</td>
            <td>${m.title}</td>
            <td>${m.planned_at ? new Date(m.planned_at).toLocaleString('tr-TR') : m.started_at ? new Date(m.started_at).toLocaleString('tr-TR') : '—'}</td>
            <td>${statusLabel[m.status] || m.status}</td>
            <td>
              ${m.status === 'planned'
                ? `<button class="btn-primary" style="width:auto;padding:0.25rem 0.75rem;font-size:0.8rem" onclick="startOturum(${m.id})">Başlat</button>`
                : ''}
              ${m.status === 'active'
                ? `<button class="btn-danger" style="width:auto;padding:0.25rem 0.75rem;font-size:0.8rem" onclick="endOturum(${m.id})">Kapat</button>`
                : ''}
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>`;
}

async function createOturum() {
  const title     = document.getElementById('oturum-baslik').value.trim();
  const meetingNo = document.getElementById('oturum-no').value.trim();
  const plannedAt = document.getElementById('oturum-tarih').value;

  if (!title) { alert('Lütfen oturum başlığı girin.'); return; }

  const res = await fetch(`${API}/meetings`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title, meeting_no: meetingNo, planned_at: plannedAt }),
  });

  if (res.ok) {
    document.getElementById('oturum-baslik').value = '';
    document.getElementById('oturum-no').value     = '';
    document.getElementById('oturum-tarih').value  = '';
    loadOturumlar();
  } else {
    const data = await res.json();
    alert(data.error || 'Oturum oluşturulamadı.');
  }
}

async function startOturum(id) {
  if (!confirm('Oturumu başlatmak istediğinize emin misiniz?')) return;
  const res = await fetch(`${API}/meetings/${id}/start`, {
    method: 'POST',
    credentials: 'include',
  });
  const data = await res.json();
  if (res.ok) { loadOturumlar(); loadPanel(); }
  else { alert(data.error || 'Oturum başlatılamadı.'); }
}

async function endOturum(id) {
  if (!confirm('Oturumu kapatmak istediğinize emin misiniz?')) return;
  const res = await fetch(`${API}/meetings/${id}/end`, {
    method: 'POST',
    credentials: 'include',
  });
  const data = await res.json();
  if (res.ok) { loadOturumlar(); loadPanel(); }
  else { alert(data.error || 'Oturum kapatılamadı.'); }
}

setInterval(() => { loadPanel(); refreshOylama(); }, 5000);
