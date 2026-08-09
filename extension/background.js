// NetView 收藏助手 - 后台服务
const CONFIG_KEY = 'netview_config';

// ---------- 配置与登录 ----------
function getConfig() {
  return new Promise((resolve) => {
    chrome.storage.sync.get(CONFIG_KEY, (r) => resolve(r[CONFIG_KEY] || {}));
  });
}

async function saveConfig(cfg) {
  await chrome.storage.sync.set({ [CONFIG_KEY]: cfg });
}

async function ensureToken(cfg) {
  if (cfg.token) return cfg.token;
  const base = (cfg.serverUrl || '').replace(/\/+$/, '');
  if (!base) throw new Error('未配置服务器地址，请先到扩展设置填写');
  if (!cfg.password) throw new Error('未配置密码，请先到扩展设置填写');
  const res = await fetch(`${base}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: cfg.password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `登录失败 (${res.status})`);
  }
  const data = await res.json();
  cfg.token = data.token;
  cfg.expires = data.expires;
  await saveConfig(cfg);
  return cfg.token;
}

async function api(cfg, path, options = {}) {
  const base = (cfg.serverUrl || '').replace(/\/+$/, '');
  const token = await ensureToken(cfg);
  const res = await fetch(`${base}${path}`, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      ...(options.headers || {}),
    },
  });
  return res;
}

// ---------- 保存动作 ----------
async function saveLink(cfg, url, type, title, description) {
  const res = await api(cfg, '/api/items', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ source_url: url, type, title: title || '', description: description || '' }),
  });
  if (!res.ok) {
    const e = await res.json().catch(() => ({}));
    throw new Error(e.error || `保存失败 (${res.status})`);
  }
  return res.json();
}

// 抓取原图片字节并直接上传到 NetView（真正把文件保存下来）
async function saveImageBytes(cfg, imageUrl, title) {
  const base = (cfg.serverUrl || '').replace(/\/+$/, '');
  const token = await ensureToken(cfg);
  const imgRes = await fetch(imageUrl, { credentials: 'include' });
  if (!imgRes.ok) throw new Error(`无法下载原图 (${imgRes.status})，可改用“保存链接”`);
  const blob = await imgRes.blob();
  const fd = new FormData();
  const clean = (imageUrl.split('?')[0].split('#')[0]);
  const ext = (clean.split('.').pop() || 'bin').slice(0, 8);
  fd.append('file', blob, `image.${ext}`);
  fd.append('type', 'image');
  fd.append('title', title || '');
  const res = await fetch(`${base}/api/items/upload`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: fd,
  });
  if (!res.ok) {
    const e = await res.json().catch(() => ({}));
    throw new Error(e.error || `上传失败 (${res.status})`);
  }
  return res.json();
}

function detectType(url) {
  const lower = (url || '').toLowerCase();
  const imgExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.avif'];
  if (imgExts.some((e) => lower.includes(e))) return 'image';
  return 'video';
}

function notify(title, message) {
  chrome.notifications.create({
    type: 'basic',
    iconUrl: 'icons/icon128.png',
    title,
    message,
  });
}

// ---------- 右键菜单 ----------
function setupMenus() {
  chrome.contextMenus.removeAll(() => {
    chrome.contextMenus.create({ id: 'save-image', title: '保存图片到 NetView', contexts: ['image'] });
    chrome.contextMenus.create({ id: 'save-image-download', title: '下载原图到 NetView（保存文件）', contexts: ['image'] });
    chrome.contextMenus.create({ id: 'save-media', title: '保存视频/媒体到 NetView', contexts: ['video', 'audio'] });
    chrome.contextMenus.create({ id: 'save-link', title: '保存此链接到 NetView', contexts: ['link'] });
    chrome.contextMenus.create({ id: 'save-page', title: '保存当前页面到 NetView', contexts: ['page'] });
  });
}

chrome.runtime.onInstalled.addListener(setupMenus);
chrome.runtime.onStartup.addListener(setupMenus);

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  try {
    const cfg = await getConfig();
    if (!cfg.serverUrl) {
      notify('未配置', '请先在扩展设置中填写 NetView 服务器地址');
      return;
    }
    const title = (tab && tab.title) || '';
    const description = info.pageUrl || '';
    switch (info.menuItemId) {
      case 'save-image':
        await saveLink(cfg, info.srcUrl, 'image', title, description);
        notify('已保存', '图片链接已加入 NetView 媒体库');
        break;
      case 'save-image-download':
        await saveImageBytes(cfg, info.srcUrl, title);
        notify('已保存', '图片原图已下载到 NetView 媒体库');
        break;
      case 'save-media':
        await saveLink(cfg, info.srcUrl, 'video', title, description);
        notify('已保存', '媒体链接已加入 NetView 媒体库');
        break;
      case 'save-link':
        await saveLink(cfg, info.linkUrl, detectType(info.linkUrl), title, description);
        notify('已保存', '链接已加入 NetView 媒体库');
        break;
      case 'save-page':
        await saveLink(cfg, info.pageUrl, 'video', title, '');
        notify('已保存', '页面链接已加入 NetView 媒体库');
        break;
    }
  } catch (e) {
    notify('保存失败', String((e && e.message) || e));
  }
});

// ---------- 与弹窗通信 ----------
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  (async () => {
    try {
      switch (msg.type) {
        case 'getStatus': {
          const cfg = await getConfig();
          sendResponse({ ok: true, configured: !!cfg.serverUrl, serverUrl: cfg.serverUrl || '' });
          break;
        }
        case 'scanImages': {
          const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
          const tab = tabs[0];
          if (!tab || !tab.id) throw new Error('未找到当前标签页');
          const result = await chrome.scripting.executeScript({
            target: { tabId: tab.id },
            func: () => {
              const seen = new Set();
              const images = [];
              for (const img of document.images) {
                const src = img.currentSrc || img.src;
                if (src && src.startsWith('http') && !seen.has(src)) {
                  seen.add(src);
                  images.push({
                    src,
                    alt: img.alt || img.title || '',
                    width: img.naturalWidth,
                    height: img.naturalHeight,
                  });
                }
              }
              return { url: location.href, title: document.title, images: images.slice(0, 60) };
            },
          });
          sendResponse({ ok: true, page: result[0] && result[0].result });
          break;
        }
        case 'saveItem': {
          const cfg = await getConfig();
          const item = await saveLink(cfg, msg.url, msg.type, msg.title || '', msg.description || '');
          sendResponse({ ok: true, item });
          break;
        }
        case 'downloadItem': {
          const cfg = await getConfig();
          const item = await saveImageBytes(cfg, msg.url, msg.title || '');
          sendResponse({ ok: true, item });
          break;
        }
        case 'testConnection': {
          const cfg = await getConfig();
          await ensureToken(cfg);
          sendResponse({ ok: true });
          break;
        }
        default:
          sendResponse({ ok: false, error: '未知操作' });
      }
    } catch (e) {
      sendResponse({ ok: false, error: String((e && e.message) || e) });
    }
  })();
  return true;
});
