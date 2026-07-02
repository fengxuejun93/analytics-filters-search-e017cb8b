// ========== 全局状态 ==========
let currentView = 'list';
let currentItem = null;
let filterOptions = null;

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

const statusMap = {
    listed: '上架中', exchanged: '已置换', delisted: '已下架',
    pending: '待处理', accepted: '已接受', rejected: '已拒绝', cancelled: '已取消',
};

// ========== 导航 ==========
function navigateTo(view, data) {
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    currentView = view;
    const el = document.getElementById('view-' + view);
    if (el) el.classList.add('active');

    if (view === 'list') {
        loadItems();
        loadStats();
    } else if (view === 'detail' && data) {
        loadDetail(data);
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
            <div class="item-card" onclick="navigateTo('detail','${item.id}')">
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
    loadStats();
}

function esc(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

// ========== 货品详情 ==========
async function loadDetail(itemId) {
    try {
        const item = await api('/items/' + itemId);
        currentItem = item;
        document.getElementById('detail-title').textContent = item.title;
        document.getElementById('detail-image').textContent = '📷';
        document.getElementById('detail-meta').innerHTML = `
            <span class="tag tag-category">${esc(item.category)}</span>
            <span class="tag tag-condition">${esc(item.condition)}</span>
            <span class="tag tag-city">${esc(item.city)}</span>
            <span class="status-badge status-${item.status}">${statusMap[item.status]}</span>
            <span style="font-size:13px;color:var(--c-text2);">发布人: ${esc(item.publisher)}</span>
        `;
        document.getElementById('detail-desc').textContent = item.description || '暂无描述';
        document.getElementById('detail-exchange').textContent = item.expected_exchange || '不限';

        // 操作按钮
        const actions = document.getElementById('detail-actions');
        let btns = '';
        if (item.status === 'listed') {
            btns += `<button class="btn btn-sm btn-primary" onclick="showEditForm('${item.id}')">编辑</button>`;
            btns += `<button class="btn btn-sm btn-secondary" onclick="delistItem('${item.id}')">下架</button>`;
        } else if (item.status === 'delisted') {
            btns += `<button class="btn btn-sm btn-success" onclick="relistItem('${item.id}')">重新上架</button>`;
        }
        actions.innerHTML = btns;

        // 申请按钮可见性
        const applyToggle = document.getElementById('btn-apply-toggle');
        applyToggle.style.display = item.status === 'listed' ? '' : 'none';
        document.getElementById('apply-form').style.display = 'none';

        loadApplications(itemId);
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 置换申请 ==========
async function loadApplications(itemId) {
    try {
        const apps = await api('/items/' + itemId + '/applications');
        const list = document.getElementById('application-list');
        if (!apps.length) {
            list.innerHTML = '<p style="color:var(--c-text2);font-size:13px;">暂无置换申请</p>';
            return;
        }
        list.innerHTML = apps.map(a => {
            let actionBtns = '';
            if (a.status === 'pending') {
                actionBtns = `
                    <button class="btn btn-sm btn-success" onclick="handleApp('${a.id}','accept')">接受</button>
                    <button class="btn btn-sm btn-danger" onclick="handleApp('${a.id}','reject')">拒绝</button>
                    <button class="btn btn-sm btn-secondary" onclick="handleApp('${a.id}','cancel')">取消</button>
                `;
            } else if (a.status === 'accepted') {
                actionBtns = `<button class="btn btn-sm btn-warning" onclick="handleApp('${a.id}','cancel')">取消置换</button>`;
            } else {
                actionBtns = `<span style="font-size:12px;color:var(--c-text2);">${statusMap[a.status]}，不可操作</span>`;
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
                    <div class="app-time">${formatTime(a.created_at)}</div>
                    <div class="app-actions">${actionBtns}</div>
                </div>
            `;
        }).join('');
    } catch (e) { showToast(e.message, 'error'); }
}

function showApplyForm() {
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
        loadApplications(currentItem.id);
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
        } else {
            await api('/items', { method: 'POST', body: JSON.stringify(payload) });
            showToast('发布成功');
        }
        navigateTo('list');
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    loadFilterOptions();
    loadItems();
    loadStats();
});
