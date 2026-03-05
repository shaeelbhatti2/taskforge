const API_KEY = localStorage.getItem('taskforge_api_key') || 'dev-api-key-change-me';

async function api(path) {
  const res = await fetch('/api/v1' + path, {
    headers: { 'X-API-Key': API_KEY, 'X-Namespace-ID': 'default' }
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

async function loadDashboard() {
  try {
    const stats = await api('/dashboard');
    document.getElementById('running').textContent = 'Running: ' + stats.running;
    document.getElementById('queued').textContent = 'Queued: ' + stats.queued;
    document.getElementById('failed').textContent = 'Failed: ' + stats.failed;
    document.getElementById('success').textContent = 'Success: ' + stats.success;
    const workers = await api('/workers');
    document.getElementById('workers').textContent = JSON.stringify(workers, null, 2);
  } catch (e) {
    document.getElementById('workers').textContent = String(e);
  }
}

async function loadRun() {
  const id = document.getElementById('runId').value;
  const run = await api('/runs/' + id);
  document.getElementById('stdout').textContent = run.stdout || '';
  document.getElementById('stderr').textContent = run.stderr || '';
}

async function loadDLQ() {
  const items = await api('/dlq');
  const tbody = document.querySelector('#dlq tbody');
  tbody.innerHTML = '';
  for (const item of items) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td>${item.id}</td><td>${item.job_id}</td><td>${item.reason}</td><td>${item.created_at}</td>`;
    tbody.appendChild(tr);
  }
}
