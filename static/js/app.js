// ========== 全局状态 ==========
let currentView = 'list';
let currentItem = null;       // 当前详情的货品对象
let currentItemId = null;     // 当前详情的货品ID
let savedFilters = {};        // 进入详情前保存的筛选状态
let filterOptions = null;
let debounceTimer = null;

// ========== 工具函数 ==========
const API = '/api';

async function api(path, options = {}) {
    const resp = await fetch(API + path, {
        headers: { 'Content-Type': 'application/json' },
        ...options,
    });
    const data = await resp.json();
    if (!resp.ok) {
        throw new Error(data.error || '请求失败');
    }
    return data;
}

function showToast(msg, type = 'success') {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.className = 'toast toast-' + type;
    t.style.display = 'block';
    clearTimeout(t._timer);
    t._timer = setTimeout(() => { t.style.display = 'none'; }, 2500);
}

function formatTime(ts) {
    const d = new Date(ts);
    const now = new Date();
    const diff = now - d;
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前';
    if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前';
    if (diff < 604800000) return Math.floor(diff / 86400000) + ' 天前';
    return d.toLocaleDateString('zh-CN');
}

function formatTimeFull(ts) {
    const d = new Date(ts);
    return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}

const statusMap = {
    listed: '上架中', exchanged: '已置换', delisted: '已下架',
    pending: '待处理', accepted: '已接受', rejected: '已拒绝', cancelled: '已取消',
};

function esc(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

function debounceLoadItems() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(loadItems, 300);
}

// ========== 导航 ==========
function navigateTo(view, data) {
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    currentView = view;
    const el = document.getElementById('view-' + view);
    if (el) el.classList.add('active');

    if (view === 'list') {
        currentItemId = null;
        currentItem = null;
        restoreFilterState();
        loadItems();
        loadStats();
    } else if (view === 'detail' && data) {
        currentItemId = data;
        loadDetail(data);
        loadStats();
    }
}

// ========== 统计 ==========
async function loadStats() {
    try {
        const stats = await api('/stats');
        document.getElementById('stat-listed').textContent = stats.listed_count;
        document.getElementById('stat-pending').textContent = stats.pending_count;
        document.getElementById('stat-exchanged').textContent = stats.exchanged_count;
        document.getElementById('stat-delisted').textContent = stats.delisted_count;
    } catch (e) { console.error(e); }
}

// ========== 筛选器选项 ==========
async function loadFilterOptions() {
    try {
        filterOptions = await api('/filters');
        const catSel = document.getElementById('filter-category');
        catSel.innerHTML = '<option value="">全部品类</option>';
        filterOptions.categories.forEach(c => {
            catSel.innerHTML += `<option value="${c}">${c}</option>`;
        });
        const citySel = document.getElementById('filter-city');
        citySel.innerHTML = '<option value="">全部城市</option>';
        filterOptions.cities.forEach(c => {
            citySel.innerHTML += `<option value="${c}">${c}</option>`;
        });
    } catch (e) { console.error(e); }
}

function saveFilterState() {
    savedFilters = {
        keyword: document.getElementById('filter-keyword').value,
        category: document.getElementById('filter-category').value,
        city: document.getElementById('filter-city').value,
        status: document.getElementById('filter-status').value,
    };
}

function restoreFilterState() {
    if (savedFilters.keyword !== undefined) {
        document.getElementById('filter-keyword').value = savedFilters.keyword;
        document.getElementById('filter-category').value = savedFilters.category;
        document.getElementById('filter-city').value = savedFilters.city;
        document.getElementById('filter-status').value = savedFilters.status;
    }
}

function clearFilters() {
    document.getElementById('filter-keyword').value = '';
    document.getElementById('filter-category').value = '';
    document.getElementById('filter-city').value = '';
    document.getElementById('filter-status').value = '';
    loadItems();
}

// ========== 货品列表 ==========
async function loadItems() {
    const params = new URLSearchParams();
    const kw = document.getElementById('filter-keyword').value.trim();
    const cat = document.getElementById('filter-category').value;
    const city = document.getElementById('filter-city').value;
    const status = document.getElementById('filter-status').value;
    if (kw) params.set('keyword', kw);
    if (cat) params.set('category', cat);
    if (city) params.set('city', city);
    if (status) params.set('status', status);

    try {
        const items = await api('/items?' + params.toString());
        const grid = document.getElementById('item-grid');
        const empty = document.getElementById('empty-state');

        if (!items.length) {
            grid.innerHTML = '';
            empty.style.display = 'block';
            return;
        }
        empty.style.display = 'none';
        grid.innerHTML = items.map(item => `
            <div class="item-card${currentItemId === item.id ? ' item-card-active' : ''}" onclick="openDetail('${item.id}')">
                <div class="item-card-image">📷</div>
                <div class="item-card-body">
                    <div class="item-card-title">${esc(item.title)}</div>
                    <div class="item-card-tags">
                        <span class="tag tag-category">${esc(item.category)}</span>
                        <span class="tag tag-condition">${esc(item.condition)}</span>
                        <span class="tag tag-city">${esc(item.city)}</span>
                    </div>
                    <div class="item-card-bottom">
                        <span class="item-card-exchange">期望: ${esc(item.expected_exchange || '不限')}</span>
                        <span class="status-badge status-${item.status}">${statusMap[item.status] || item.status}</span>
                    </div>
                    <div class="item-card-bottom" style="margin-top:6px;">
                        <span>${esc(item.publisher)}</span>
                        <span>${formatTime(item.updated_at)}</span>
                    </div>
                </div>
            </div>
        `).join('');
    } catch (e) { showToast(e.message, 'error'); }
}

function openDetail(itemId) {
    // 进入详情前保存筛选状态，返回时恢复
    saveFilterState();
    navigateTo('detail', itemId);
}

// ========== 货品详情（聚合接口） ==========
async function loadDetail(itemId) {
    try {
        const detail = await api('/items/' + itemId + '/detail');
        currentItem = detail.item;
        currentItemId = detail.item.id;
        renderDetail(detail);
    } catch (e) { showToast(e.message, 'error'); }
}

function renderDetail(detail) {
    const item = detail.item;
    const apps = detail.applications || [];
    const history = detail.status_history || [];

    // 导航面包屑
    document.getElementById('detail-nav-item').textContent = item.title;

    // 图片占位
    document.getElementById('detail-image').textContent = '📷';

    // 标题
    document.getElementById('detail-title').textContent = item.title;

    // 元信息标签
    document.getElementById('detail-meta').innerHTML = `
        <span class="tag tag-category">${esc(item.category)}</span>
        <span class="tag tag-condition">${esc(item.condition)}</span>
        <span class="tag tag-city">${esc(item.city)}</span>
        <span class="status-badge status-${item.status}">${statusMap[item.status]}</span>
    `;

    // 字段表格
    document.getElementById('detail-fields').innerHTML = `
        <div class="field-row"><span class="field-label">品类</span><span class="field-value">${esc(item.category)}</span></div>
        <div class="field-row"><span class="field-label">成色</span><span class="field-value">${esc(item.condition)}</span></div>
        <div class="field-row"><span class="field-label">所在城市</span><span class="field-value">${esc(item.city)}</span></div>
        <div class="field-row"><span class="field-label">发布人</span><span class="field-value">${esc(item.publisher)}</span></div>
        <div class="field-row"><span class="field-label">期望置换物</span><span class="field-value">${esc(item.expected_exchange || '不限')}</span></div>
        <div class="field-row"><span class="field-label">发布时间</span><span class="field-value">${formatTimeFull(item.created_at)}</span></div>
        <div class="field-row"><span class="field-label">最后更新</span><span class="field-value">${formatTimeFull(item.updated_at)}</span></div>
    `;

    // 描述与交换条件
    document.getElementById('detail-desc').textContent = item.description || '暂无描述';
    document.getElementById('detail-exchange').textContent = item.expected_exchange || '不限';

    // 操作按钮 + 状态提示
    renderDetailActions(item);

    // 申请列表
    renderApplications(apps, item);

    // 状态历史时间线
    renderStatusHistory(history);
}

function renderDetailActions(item) {
    const actions = document.getElementById('detail-actions');
    const hint = document.getElementById('detail-status-hint');
    let btns = '';
    let hintText = '';

    switch (item.status) {
        case 'listed':
            btns += `<button class="btn btn-sm btn-primary" onclick="showEditForm('${item.id}')">编辑</button>`;
            btns += `<button class="btn btn-sm btn-secondary" onclick="delistItem('${item.id}')">下架</button>`;
            break;
        case 'delisted':
            btns += `<button class="btn btn-sm btn-success" onclick="relistItem('${item.id}')">重新上架</button>`;
            btns += `<button class="btn btn-sm btn-primary" onclick="showEditForm('${item.id}')">编辑</button>`;
            hintText = '此货品已下架，无法发起新的置换申请。可重新上架或编辑后恢复。';
            break;
        case 'exchanged':
            hintText = '此货品已达成置换，不可编辑、下架或发起新申请。如需恢复，请取消已接受的置换申请。';
            break;
    }

    actions.innerHTML = btns;

    if (hintText) {
        hint.innerHTML = hintText;
        hint.style.display = 'flex';
    } else {
        hint.style.display = 'none';
    }
}

// ========== 置换申请渲染 ==========
function renderApplications(apps, item) {
    // 申请数量 badge
    document.getElementById('detail-app-count').textContent = apps.length;

    // 申请按钮可见性
    const applyToggle = document.getElementById('btn-apply-toggle');
    if (item.status === 'listed') {
        applyToggle.style.display = '';
    } else {
        applyToggle.style.display = 'none';
    }
    document.getElementById('apply-form').style.display = 'none';

    const list = document.getElementById('application-list');
    if (!apps.length) {
        list.innerHTML = '<p class="empty-hint">暂无置换申请</p>';
        return;
    }

    list.innerHTML = apps.map(a => {
        let actionBtns = '';
        // 货品已下架/已置换时，申请也不可操作
        const canOperate = item.status === 'listed' || item.status === 'exchanged';

        if (a.status === 'pending' && item.status === 'listed') {
            actionBtns = `
                <button class="btn btn-sm btn-success" onclick="handleApp('${a.id}','accept')">接受</button>
                <button class="btn btn-sm btn-danger" onclick="handleApp('${a.id}','reject')">拒绝</button>
                <button class="btn btn-sm btn-secondary" onclick="handleApp('${a.id}','cancel')">取消</button>
            `;
        } else if (a.status === 'accepted') {
            actionBtns = `<button class="btn btn-sm btn-warning" onclick="handleApp('${a.id}','cancel')">取消置换</button>`;
        } else if (a.status === 'pending' && item.status !== 'listed') {
            actionBtns = `<span class="op-hint">货品已${statusMap[item.status]}，暂不可操作</span>`;
        } else {
            actionBtns = `<span class="op-hint">${statusMap[a.status]}，不可操作</span>`;
        }

        const statusColors = { pending: '#d97706', accepted: '#16a34a', rejected: '#dc2626', cancelled: '#64748b' };
        const statusBgs = { pending: '#fef3c7', accepted: '#dcfce7', rejected: '#fee2e2', cancelled: '#f1f5f9' };

        return `
            <div class="app-card app-${a.status}">
                <div class="app-header-row">
                    <span class="app-applicant">${esc(a.applicant)}</span>
                    <span class="app-status-badge" style="color:${statusColors[a.status]};background:${statusBgs[a.status]}">${statusMap[a.status]}</span>
                </div>
                <div class="app-message">${esc(a.message)}</div>
                <div class="app-time">申请于 ${formatTime(a.created_at)}${a.updated_at !== a.created_at ? ' · 更新于 ' + formatTime(a.updated_at) : ''}</div>
                <div class="app-actions">${actionBtns}</div>
            </div>
        `;
    }).join('');
}

// ========== 状态历史时间线 ==========
function renderStatusHistory(histories) {
    const el = document.getElementById('status-history');
    if (!histories || !histories.length) {
        el.innerHTML = '<p class="empty-hint">暂无状态变更记录</p>';
        return;
    }

    el.innerHTML = '<div class="timeline">' + histories.map((h, i) => {
        const isLast = i === histories.length - 1;
        const fromLabel = h.from_status ? statusMap[h.from_status] : '新建';
        const toLabel = statusMap[h.to_status] || h.to_status;
        return `
            <div class="timeline-item${isLast ? ' timeline-item-active' : ''}">
                <div class="timeline-dot"></div>
                <div class="timeline-content">
                    <div class="timeline-title">${esc(fromLabel)} → ${esc(toLabel)}</div>
                    <div class="timeline-reason">${esc(h.reason)}</div>
                    <div class="timeline-meta">操作人: ${esc(h.operator)} · ${formatTime(h.created_at)}</div>
                </div>
            </div>
        `;
    }).join('') + '</div>';
}

// ========== 申请交互 ==========
function showApplyForm() {
    if (!currentItem || currentItem.status !== 'listed') {
        showToast('该货品当前状态不允许发起申请', 'error');
        return;
    }
    const f = document.getElementById('apply-form');
    f.style.display = f.style.display === 'none' ? 'flex' : 'none';
}

async function submitApplication() {
    const applicant = document.getElementById('apply-applicant').value.trim();
    const message = document.getElementById('apply-message').value.trim();
    if (!applicant) { showToast('请填写申请人名字', 'error'); return; }
    try {
        await api('/items/' + currentItem.id + '/applications', {
            method: 'POST',
            body: JSON.stringify({ applicant, message }),
        });
        showToast('申请已提交');
        document.getElementById('apply-applicant').value = '';
        document.getElementById('apply-message').value = '';
        document.getElementById('apply-form').style.display = 'none';
        // 重新加载完整详情保证数据一致
        loadDetail(currentItem.id);
        loadStats();
    } catch (e) { showToast(e.message, 'error'); }
}

async function handleApp(appId, action) {
    try {
        await api('/applications/' + appId, {
            method: 'PUT',
            body: JSON.stringify({ action }),
        });
        const actionLabels = { accept: '已接受', reject: '已拒绝', cancel: '已取消' };
        showToast(actionLabels[action] || '操作成功');
        // 重新加载完整详情（状态、申请、历史全部刷新）
        loadDetail(currentItem.id);
        loadStats();
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 货品状态操作 ==========
async function delistItem(id) {
    try {
        await api('/items/' + id + '/status', {
            method: 'PUT',
            body: JSON.stringify({ status: 'delisted' }),
        });
        showToast('已下架');
        loadDetail(id);
        loadStats();
    } catch (e) { showToast(e.message, 'error'); }
}

async function relistItem(id) {
    try {
        await api('/items/' + id + '/status', {
            method: 'PUT',
            body: JSON.stringify({ status: 'listed' }),
        });
        showToast('已重新上架');
        loadDetail(id);
        loadStats();
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 创建/编辑表单 ==========
function showCreateForm() {
    document.getElementById('form-title').textContent = '发布新货品';
    document.getElementById('form-item-id').value = '';
    document.getElementById('item-form').reset();
    // 新建返回列表
    document.getElementById('form-back-btn').onclick = () => navigateTo('list');
    document.getElementById('form-cancel-btn').onclick = () => navigateTo('list');
    navigateTo('form');
}

function showEditForm(id) {
    const item = currentItem;
    document.getElementById('form-title').textContent = '编辑货品';
    document.getElementById('form-item-id').value = item.id;
    document.getElementById('form-field-title').value = item.title;
    document.getElementById('form-field-category').value = item.category;
    document.getElementById('form-field-condition').value = item.condition;
    document.getElementById('form-field-city').value = item.city;
    document.getElementById('form-field-exchange').value = item.expected_exchange;
    document.getElementById('form-field-publisher').value = item.publisher;
    document.getElementById('form-field-description').value = item.description;
    // 编辑返回详情
    document.getElementById('form-back-btn').onclick = () => navigateTo('detail', item.id);
    document.getElementById('form-cancel-btn').onclick = () => navigateTo('detail', item.id);
    navigateTo('form');
}

async function submitItemForm(e) {
    e.preventDefault();
    const id = document.getElementById('form-item-id').value;
    const payload = {
        title: document.getElementById('form-field-title').value.trim(),
        category: document.getElementById('form-field-category').value,
        condition: document.getElementById('form-field-condition').value,
        city: document.getElementById('form-field-city').value.trim(),
        expected_exchange: document.getElementById('form-field-exchange').value.trim(),
        publisher: document.getElementById('form-field-publisher').value.trim(),
        description: document.getElementById('form-field-description').value.trim(),
    };

    try {
        if (id) {
            await api('/items/' + id, { method: 'PUT', body: JSON.stringify(payload) });
            showToast('编辑成功');
            // 编辑保存后回到该货品详情
            navigateTo('detail', id);
        } else {
            const newItem = await api('/items', { method: 'POST', body: JSON.stringify(payload) });
            showToast('发布成功');
            // 新建后回到该货品详情
            navigateTo('detail', newItem.id);
        }
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    // 显示访问入口
    const urlEl = document.getElementById('header-url');
    urlEl.textContent = window.location.href;

    loadFilterOptions();
    loadItems();
    loadStats();
});
