const API = '/api';
let token = localStorage.getItem('token') || '';
let currentUser = null;

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

async function api(method, path, body) {
    const opts = {
        method,
        headers: { 'Content-Type': 'application/json' },
    };
    if (token) opts.headers['Authorization'] = 'Bearer ' + token;
    if (body) opts.body = JSON.stringify(body);

    const res = await fetch(API + path, opts);
    if (res.status === 204) return null;
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Request failed');
    return data;
}

function toast(msg, type = 'success') {
    const el = document.createElement('div');
    el.className = 'toast ' + type;
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(() => el.remove(), 3000);
}

// Auth
function initAuth() {
    if (token) {
        showApp();
        return;
    }
    $('#auth-screen').style.display = '';
    $('#app-screen').style.display = 'none';
}

$$('.auth-tabs button').forEach((btn) => {
    btn.addEventListener('click', () => {
        $$('.auth-tabs button').forEach((b) => b.classList.remove('active'));
        btn.classList.add('active');
        const mode = btn.dataset.mode;
        $('#register-fields').style.display = mode === 'register' ? '' : 'none';
        $('#auth-submit').textContent = mode === 'register' ? 'Зарегистрироваться' : 'Войти';
        $('#auth-form').dataset.mode = mode;
    });
});

$('#auth-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const mode = e.target.dataset.mode;
    const login = $('#auth-login').value;
    const password = $('#auth-password').value;

    try {
        if (mode === 'register') {
            const email = $('#auth-email').value;
            await api('POST', '/auth/register', { login, email, password });
            toast('Регистрация успешна!');
        }
        const data = await api('POST', '/auth/login', { login, password });
        token = data.token;
        localStorage.setItem('token', token);
        showApp();
    } catch (err) {
        toast(err.message, 'error');
    }
});

function showApp() {
    $('#auth-screen').style.display = 'none';
    $('#app-screen').style.display = '';
    parseToken();
    navigate('workouts');
}

function parseToken() {
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        currentUser = { id: payload.user_id, role: payload.role };
        $('.user-name').textContent = currentUser.role === 'admin' ? 'Admin' : 'User #' + currentUser.id;
    } catch {
        logout();
    }
}

function logout() {
    token = '';
    currentUser = null;
    localStorage.removeItem('token');
    $('#auth-screen').style.display = '';
    $('#app-screen').style.display = 'none';
}

$('#logout-btn').addEventListener('click', logout);

// Navigation
function navigate(page) {
    $$('.page').forEach((p) => p.classList.remove('active'));
    $$('.nav-links button').forEach((b) => b.classList.remove('active'));
    $(`#page-${page}`).classList.add('active');
    $(`.nav-links button[data-page="${page}"]`).classList.add('active');

    const loaders = {
        workouts: loadWorkouts,
        exercises: loadExercises,
        stats: loadStats,
    };
    if (loaders[page]) loaders[page]();
}

$$('.nav-links button').forEach((btn) => {
    btn.addEventListener('click', () => navigate(btn.dataset.page));
});

// Workouts
async function loadWorkouts() {
    try {
        const workouts = await api('GET', '/workouts');
        const list = $('#workouts-list');
        if (!workouts || workouts.length === 0) {
            list.innerHTML = '<div class="card"><p>Нет тренировок. Создайте первую!</p></div>';
            return;
        }
        list.innerHTML = workouts.map((w) => `
            <div class="card">
                <div class="card-header">
                    <div>
                        <div class="card-title">${esc(w.title)}</div>
                        <div class="card-subtitle">${formatDate(w.date)} · ${w.duration_minutes} мин</div>
                    </div>
                    <div class="card-actions">
                        <button class="btn btn-sm btn-outline" onclick="copyWorkout(${w.id})">Копировать</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteWorkout(${w.id})">Удалить</button>
                    </div>
                </div>
                ${w.notes ? '<p style="color:var(--text-muted);font-size:0.85rem">' + esc(w.notes) + '</p>' : ''}
            </div>
        `).join('');
    } catch (err) {
        toast(err.message, 'error');
    }
}

$('#btn-new-workout').addEventListener('click', () => {
    $('#workout-modal').classList.add('active');
    loadExerciseOptions();
});

$('#workout-modal .modal-overlay, #close-workout-modal').forEach?.(() => {});
$('#close-workout-modal').addEventListener('click', () => {
    $('#workout-modal').classList.remove('active');
});

$('#workout-modal').addEventListener('click', (e) => {
    if (e.target === $('#workout-modal')) $('#workout-modal').classList.remove('active');
});

let exerciseRows = 0;

async function loadExerciseOptions() {
    try {
        const exercises = await api('GET', '/exercises');
        window._exercises = exercises || [];
        $('#workout-exercises').innerHTML = '';
        exerciseRows = 0;
        addExerciseRow();
    } catch (err) {
        toast(err.message, 'error');
    }
}

function addExerciseRow() {
    const container = $('#workout-exercises');
    const opts = (window._exercises || [])
        .map((e) => `<option value="${e.id}">${esc(e.name)} (${esc(e.muscle_group)})</option>`)
        .join('');

    const row = document.createElement('div');
    row.className = 'exercise-row';
    row.innerHTML = `
        <select name="exercise_id">${opts}</select>
        <input type="number" name="sets" placeholder="Подходы" min="1" value="3">
        <input type="number" name="reps" placeholder="Повторы" min="1" value="10">
        <input type="number" name="weight" placeholder="Вес (кг)" min="0" step="0.5" value="0">
        <button class="btn btn-sm btn-danger" onclick="this.parentElement.remove()">x</button>
    `;
    container.appendChild(row);
    exerciseRows++;
}

$('#add-exercise-row').addEventListener('click', addExerciseRow);

$('#workout-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = $('#w-title').value;
    const date = $('#w-date').value || new Date().toISOString().slice(0, 10);
    const duration = parseInt($('#w-duration').value) || 0;
    const notes = $('#w-notes').value;

    const exercises = [];
    $$('#workout-exercises .exercise-row').forEach((row) => {
        exercises.push({
            exercise_id: parseInt(row.querySelector('[name=exercise_id]').value),
            sets: parseInt(row.querySelector('[name=sets]').value) || 0,
            reps: parseInt(row.querySelector('[name=reps]').value) || 0,
            weight_kg: parseFloat(row.querySelector('[name=weight]').value) || 0,
        });
    });

    try {
        await api('POST', '/workouts', { title, date, duration_minutes: duration, notes, exercises });
        $('#workout-modal').classList.remove('active');
        toast('Тренировка сохранена!');
        loadWorkouts();
    } catch (err) {
        toast(err.message, 'error');
    }
});

async function deleteWorkout(id) {
    if (!confirm('Удалить тренировку?')) return;
    try {
        await api('DELETE', '/workouts/' + id);
        toast('Тренировка удалена');
        loadWorkouts();
    } catch (err) {
        toast(err.message, 'error');
    }
}

async function copyWorkout(id) {
    try {
        await api('POST', '/workouts/' + id + '/copy');
        toast('Тренировка скопирована!');
        loadWorkouts();
    } catch (err) {
        toast(err.message, 'error');
    }
}

// Exercises
async function loadExercises() {
    try {
        const exercises = await api('GET', '/exercises');
        const list = $('#exercises-list');

        const adminBtns = currentUser && currentUser.role === 'admin'
            ? '<button class="btn btn-sm btn-danger" onclick="deleteExercise(${e.id})">Удалить</button>'
            : '';

        if (!exercises || exercises.length === 0) {
            list.innerHTML = '<div class="card"><p>Каталог упражнений пуст.</p></div>';
        } else {
            list.innerHTML = `<table>
                <thead><tr><th>ID</th><th>Название</th><th>Группа мышц</th><th>Описание</th>${currentUser?.role === 'admin' ? '<th></th>' : ''}</tr></thead>
                <tbody>${exercises.map((e) => `
                    <tr>
                        <td>${e.id}</td>
                        <td>${esc(e.name)}</td>
                        <td><span class="badge">${esc(e.muscle_group)}</span></td>
                        <td style="color:var(--text-muted)">${esc(e.description)}</td>
                        ${currentUser?.role === 'admin' ? `<td><button class="btn btn-sm btn-danger" onclick="deleteExercise(${e.id})">x</button></td>` : ''}
                    </tr>
                `).join('')}</tbody>
            </table>`;
        }

        if (currentUser && currentUser.role === 'admin') {
            $('#admin-exercise-form').style.display = '';
        } else {
            $('#admin-exercise-form').style.display = 'none';
        }
    } catch (err) {
        toast(err.message, 'error');
    }
}

$('#exercise-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    try {
        await api('POST', '/exercises', {
            name: $('#ex-name').value,
            muscle_group: $('#ex-muscle').value,
            description: $('#ex-desc').value,
        });
        toast('Упражнение добавлено!');
        $('#ex-name').value = '';
        $('#ex-desc').value = '';
        loadExercises();
    } catch (err) {
        toast(err.message, 'error');
    }
});

async function deleteExercise(id) {
    if (!confirm('Удалить упражнение?')) return;
    try {
        await api('DELETE', '/exercises/' + id);
        toast('Упражнение удалено');
        loadExercises();
    } catch (err) {
        toast(err.message, 'error');
    }
}

// Stats
async function loadStats() {
    try {
        const [prData, volumeData] = await Promise.all([
            api('GET', '/stats/pr'),
            api('GET', '/stats/volume'),
        ]);

        $('#stat-volume').textContent = Math.round(volumeData.weekly_volume).toLocaleString() + ' кг';

        const exercises = window._exercises || [];
        const prList = $('#pr-list');

        if (!prData || prData.length === 0) {
            prList.innerHTML = '<p style="color:var(--text-muted)">Пока нет рекордов</p>';
            $('#stat-pr-count').textContent = '0';
            return;
        }

        $('#stat-pr-count').textContent = prData.length;
        prList.innerHTML = `<table>
            <thead><tr><th>Упражнение</th><th>Вес</th><th>Подходы x Повторы</th></tr></thead>
            <tbody>${prData.map((r) => {
                const ex = exercises.find((e) => e.id === r.exercise_id);
                const name = ex ? ex.name : '#' + r.exercise_id;
                return `<tr>
                    <td>${esc(name)}</td>
                    <td><span class="badge badge-pr">${r.weight_kg} кг</span></td>
                    <td>${r.sets}x${r.reps}</td>
                </tr>`;
            }).join('')}</tbody>
        </table>`;
    } catch (err) {
        toast(err.message, 'error');
    }
}

// Helpers
function esc(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

function formatDate(s) {
    const d = new Date(s);
    return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

// Init
initAuth();
