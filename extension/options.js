const CONFIG_KEY = 'netview_config';
const statusEl = document.getElementById('status');
const serverInput = document.getElementById('serverUrl');
const pwdInput = document.getElementById('password');

function setStatus(html, cls) {
  statusEl.innerHTML = html;
  statusEl.className = 'status ' + (cls || '');
}

chrome.storage.sync.get(CONFIG_KEY, (r) => {
  const cfg = r[CONFIG_KEY] || {};
  if (cfg.serverUrl) serverInput.value = cfg.serverUrl;
  if (cfg.password) pwdInput.value = cfg.password;
});

function saveCfg() {
  return new Promise((resolve) => {
    chrome.storage.sync.set({ [CONFIG_KEY]: { serverUrl: serverInput.value.trim(), password: pwdInput.value } }, resolve);
  });
}

function send(msg) {
  return new Promise((resolve) => chrome.runtime.sendMessage(msg, (r) => resolve(r || { ok: false, error: '无响应' })));
}

document.getElementById('save').addEventListener('click', async () => {
  if (!serverInput.value.trim()) {
    setStatus('请填写服务器地址', 'err');
    return;
  }
  await saveCfg();
  const r = await send({ type: 'testConnection' });
  if (r.ok) {
    setStatus('配置已保存，连接正常 ✓', 'ok');
  } else {
    setStatus('配置已保存，但连接失败：' + (r.error || ''), 'err');
  }
});

document.getElementById('test').addEventListener('click', async () => {
  await saveCfg();
  setStatus('连接中…', '');
  const r = await send({ type: 'testConnection' });
  if (r.ok) setStatus('连接正常 ✓', 'ok');
  else setStatus('连接失败：' + (r.error || ''), 'err');
});
