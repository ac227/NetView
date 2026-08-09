// 弹窗：显示当前页面图片，可一键保存
const app = document.getElementById('app');

function send(msg) {
  return new Promise((resolve) => chrome.runtime.sendMessage(msg, (r) => resolve(r || { ok: false, error: '无响应' })));
}

function el(tag, props = {}, children = []) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(props)) {
    if (k === 'text') node.textContent = v;
    else if (k.startsWith('on')) node.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'style') Object.assign(node.style, v);
    else node.setAttribute(k, v);
  }
  for (const c of children) node.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
  return node;
}

async function init() {
  const status = await send({ type: 'getStatus' });

  if (!status.configured) {
    app.innerHTML = '';
    app.appendChild(el('div', { text: '还没有配置 NetView 服务器。' }));
    const url = el('input', { type: 'text', placeholder: '服务器地址，如 http://192.168.1.10:8080' });
    const pwd = el('input', { type: 'password', placeholder: '访问密码' });
    const err = el('div', { className: 'err' });
    const saveBtn = el('button', { className: 'btn', text: '保存配置', onClick: async () => {
      if (!url.value.trim()) { err.textContent = '请填写服务器地址'; return; }
      const res = await new Promise((r) => chrome.storage.sync.set({
        netview_config: { serverUrl: url.value.trim(), password: pwd.value },
      }, () => r(true)));
      const t = await send({ type: 'testConnection' });
      if (!t.ok) { err.textContent = t.error || '连接失败'; return; }
      err.textContent = '';
      init();
    } });
    app.append(el('div', { className: 'tip', text: '也可在扩展“选项”页配置。' }), url, pwd, saveBtn, err);
    return;
  }

  app.innerHTML = '';
  app.appendChild(el('div', { className: 'meta', text: `服务器：${status.serverUrl}` }));

  const pageBtn = el('button', { className: 'btn', text: '保存当前页面', onClick: async () => {
    const scan = await send({ type: 'scanImages' });
    if (!scan.ok || !scan.page) return;
    const r = await send({ type: 'saveItem', url: scan.page.url, type: 'video', title: scan.page.title || '', description: '' });
    if (r.ok) pageBtn.textContent = '✓ 已保存';
    else pageBtn.textContent = '保存失败';
  } });
  app.appendChild(pageBtn);

  const scan = await send({ type: 'scanImages' });
  if (!scan.ok) {
    app.appendChild(el('div', { className: 'err', text: scan.error || '扫描页面失败' }));
    return;
  }
  const page = scan.page;
  app.appendChild(el('div', { className: 'meta', text: `${page.title}（${page.images.length} 张图片）` }));

  if (page.images.length === 0) {
    app.appendChild(el('div', { className: 'empty', text: '当前页面没有可识别的图片' }));
    return;
  }

  for (const img of page.images) {
    const saveBtn = el('button', { className: 'btn', text: '保存', onClick: async () => {
      saveBtn.disabled = true;
      const r = await send({ type: 'saveItem', url: img.src, type: 'image', title: img.alt || page.title || '', description: page.url || '' });
      saveBtn.textContent = r.ok ? '✓' : '✗';
      saveBtn.disabled = false;
    } });
    const dlBtn = el('button', { className: 'btn gray', text: '原图', onClick: async () => {
      dlBtn.disabled = true;
      const r = await send({ type: 'downloadItem', url: img.src, title: img.alt || page.title || '' });
      dlBtn.textContent = r.ok ? '✓' : '✗';
      dlBtn.disabled = false;
    } });
    const row = el('div', { className: 'row' }, [
      el('img', { className: 'thumb', src: img.src }),
      el('div', { className: 'info' }, [
        el('div', { className: 'alt', text: img.alt || '(无描述)' }),
        el('div', { className: 'dim', text: img.width && img.height ? `${img.width}×${img.height}` : '' }),
      ]),
      dlBtn,
      saveBtn,
    ]);
    app.appendChild(row);
  }
}

init();
