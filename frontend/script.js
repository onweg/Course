const API_URL = 'http://localhost:8080/api';

let currentUser = null;
let authToken = null;

// Инициализация при загрузке
document.addEventListener('DOMContentLoaded', () => {
    console.log('Инициализация приложения...');
    
    // Обработчики форм
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', handleLogout);
    }
    
    const trainingForm = document.getElementById('training-form');
    if (trainingForm) {
        trainingForm.addEventListener('submit', handleCreateTraining);
    }
    
    const trainingType = document.getElementById('training-type');
    if (trainingType) {
        trainingType.addEventListener('change', toggleMaxParticipants);
    }
    
    const createTrainingBtn = document.getElementById('create-training-btn');
    if (createTrainingBtn) {
        createTrainingBtn.addEventListener('click', () => {
            openTrainingModal();
        });
    }
    
    console.log('Обработчики событий установлены');
    
    // Проверяем сохраненную сессию
    const savedToken = localStorage.getItem('authToken');
    const savedUser = localStorage.getItem('currentUser');
    
    if (savedToken && savedUser) {
        console.log('Найдена сохраненная сессия');
        authToken = savedToken;
        try {
            currentUser = JSON.parse(savedUser);
            console.log('Пользователь восстановлен:', currentUser);
            showApp();
        } catch (e) {
            console.error('Ошибка парсинга сохраненного пользователя:', e);
            localStorage.removeItem('authToken');
            localStorage.removeItem('currentUser');
            showLogin();
        }
    } else {
        console.log('Сохраненная сессия не найдена, показываем логин');
        showLogin();
    }
});

// Показать страницу логина
function showLogin() {
    const loginPage = document.getElementById('login-page');
    const appPage = document.getElementById('app-page');
    
    if (loginPage) {
        loginPage.classList.add('active');
    }
    
    if (appPage) {
        appPage.classList.remove('active');
    }
}

// Активировать вкладку по имени и загрузить данные
function setActiveTab(tabName) {
    // Сброс активных классов
    document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));

    const tabContent = document.getElementById(`${tabName}-tab`);
    const tabBtn = document.querySelector(`.tab-btn[data-tab="${tabName}"]`);

    if (tabContent) tabContent.classList.add('active');
    if (tabBtn) tabBtn.classList.add('active');

    // Загружаем данные для выбранной вкладки
    if (tabName === 'trainings') {
        loadTrainersForFilter();
        loadTrainings();
    } else if (tabName === 'my-trainings') {
        loadMyTrainings();
    } else if (tabName === 'users') {
        loadUsers();
    } else if (tabName === 'clients') {
        loadClients();
    } else if (tabName === 'subscriptions') {
        loadSubscriptions();
    } else if (tabName === 'employees') {
        loadEmployees();
    }
}

// Показать приложение
function showApp() {
    console.log('showApp вызвана, currentUser:', currentUser);
    
    try {
        const loginPage = document.getElementById('login-page');
        const appPage = document.getElementById('app-page');
        
        console.log('loginPage:', loginPage);
        console.log('appPage:', appPage);
        
        if (!loginPage) {
            console.error('Элемент login-page не найден!');
            return;
        }
        
        if (!appPage) {
            console.error('Элемент app-page не найден!');
            return;
        }
        
        // Переключаем страницы через классы
        loginPage.classList.remove('active');
        appPage.classList.add('active');
        
        console.log('Страницы переключены');
        
        // Обновляем информацию о пользователе
        const userNameEl = document.getElementById('user-name');
        const userRoleEl = document.getElementById('user-role');
        
        console.log('userNameEl:', userNameEl);
        console.log('userRoleEl:', userRoleEl);
        
        if (userNameEl && currentUser) {
            userNameEl.textContent = currentUser.name;
            console.log('Имя пользователя обновлено:', currentUser.name);
        }
        
        if (userRoleEl && currentUser) {
            userRoleEl.textContent = currentUser.role;
            userRoleEl.className = `badge ${currentUser.role}`;
            console.log('Роль пользователя обновлена:', currentUser.role);
        }
        
        // Показываем/скрываем вкладки в зависимости от роли
        updateTabsVisibility();
        
        // Всегда активируем первую вкладку (тренировки)
        setActiveTab('trainings');
        
        console.log('showApp завершена успешно');
    } catch (error) {
        console.error('Ошибка в showApp:', error);
        alert('Ошибка переключения страницы: ' + error.message);
    }
}

// Обновить видимость вкладок
function updateTabsVisibility() {
    if (!currentUser) {
        console.error('currentUser не установлен');
        return;
    }
    
    console.log('Обновление видимости вкладок для роли:', currentUser.role);
    
    // Обновляем видимость кнопок вкладок
    document.querySelectorAll('.tab-btn[data-role]').forEach(btn => {
        const roles = btn.getAttribute('data-role').split(',');
        const shouldShow = roles.includes(currentUser.role);
        btn.style.display = shouldShow ? 'block' : 'none';
        console.log(`Вкладка ${btn.getAttribute('data-tab')}: ${shouldShow ? 'показана' : 'скрыта'}`);
    });
    
    // Обновляем видимость других элементов с data-role
    document.querySelectorAll('[data-role]:not(.tab-btn)').forEach(el => {
        const roles = el.getAttribute('data-role').split(',');
        const shouldShow = roles.includes(currentUser.role);
        el.style.display = shouldShow ? '' : 'none';
    });
}

// Вход
async function handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    console.log('Попытка входа:', email);
    console.log('URL:', `${API_URL}/auth/login`);

    try {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        console.log('Ответ получен:', response.status, response.statusText);

        if (!response.ok) {
            const error = await response.text();
            console.error('Ошибка ответа:', error);
            alert('Ошибка входа: ' + error);
            return;
        }

        const data = await response.json();
        console.log('Данные получены:', data);
        
        if (!data.token || !data.user) {
            console.error('Неполные данные от сервера:', data);
            alert('Ошибка: неполные данные от сервера');
            return;
        }
        
        authToken = data.token;
        currentUser = data.user;

        // Сохраняем в localStorage
        localStorage.setItem('authToken', authToken);
        localStorage.setItem('currentUser', JSON.stringify(currentUser));

        console.log('Токен сохранен, переключаем на приложение...');
        
        // Очищаем форму
        const loginForm = document.getElementById('login-form');
        if (loginForm) {
            loginForm.reset();
        }
        
        // Небольшая задержка для визуального эффекта
        setTimeout(() => {
            showApp();
        }, 100);
    } catch (error) {
        console.error('Ошибка подключения:', error);
        console.error('Тип ошибки:', error.name);
        console.error('Сообщение:', error.message);
        alert('Ошибка подключения к серверу: ' + error.message);
    }
}

// Выход
function handleLogout() {
    if (authToken) {
        fetch(`${API_URL}/auth/logout`, {
            method: 'POST',
            headers: { 'Authorization': authToken }
        }).catch(console.error);
    }

    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    authToken = null;
    currentUser = null;
    showLogin();
}

// Переключение вкладок (обертка для кликов)
function showTab(tabName) {
    setActiveTab(tabName);
}

// Загрузка тренеров для фильтра
async function loadTrainersForFilter() {
    try {
        const response = await fetch(`${API_URL}/users`, {
            headers: { 'Authorization': authToken }
        });
        const users = await response.json();
        const select = document.getElementById('filter-trainer');
        if (!select) return;
        
        select.innerHTML = '<option value="">Все тренеры</option>';
        
        const trainers = users.filter(u => u.role === 'trainer' || u.role === 'admin');
        trainers.forEach(trainer => {
            const option = document.createElement('option');
            option.value = trainer.id;
            option.textContent = trainer.name;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('Ошибка загрузки тренеров для фильтра:', error);
    }
}

async function updateTrainingStatus(trainingId, selectEl) {
    const newStatus = selectEl?.value ?? '';
    const prevStatus = selectEl?.getAttribute('data-prev-status') || newStatus;

    if (!['scheduled', 'completed', 'cancelled'].includes(newStatus)) {
        alert('Недопустимый статус');
        return;
    }

    // Мини-UI: блокируем select и показываем подсказку
    let hint;
    if (selectEl) {
        selectEl.disabled = true;
        hint = selectEl.nextElementSibling;
        if (!hint || !hint.classList?.contains('status-hint')) {
            hint = document.createElement('span');
            hint.className = 'status-hint';
            selectEl.after(hint);
        }
        hint.textContent = 'Сохраняем...';
        hint.classList.remove('error', 'success');
    }

    try {
        const resp = await fetch(`${API_URL}/trainings/${trainingId}/status`, {
            method: 'PUT',
            mode: 'cors',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': authToken
            },
            body: JSON.stringify({ status: newStatus })
        });
        if (!resp.ok) {
            const error = await resp.text();
            if (selectEl) selectEl.value = prevStatus;
            if (hint) {
                hint.textContent = error || 'Ошибка сохранения';
                hint.classList.add('error');
            }
            return;
        }
        if (selectEl) selectEl.setAttribute('data-prev-status', newStatus);
        if (hint) {
            hint.textContent = 'Обновлено';
            hint.classList.add('success');
            setTimeout(() => hint.remove(), 1500);
        }
        // обновляем списки
        if (document.getElementById('trainings-tab')?.classList.contains('active')) {
            loadTrainings();
        }
        if (document.getElementById('my-trainings-tab')?.classList.contains('active')) {
            loadMyTrainings();
        }
    } catch (error) {
        console.error('Ошибка обновления статуса:', error);
        if (selectEl) selectEl.value = prevStatus;
        if (hint) {
            hint.textContent = error.message || 'Ошибка';
            hint.classList.add('error');
        }
    } finally {
        if (selectEl) selectEl.disabled = false;
    }
}

// Загрузка тренировок
async function loadTrainings() {
    try {
        const hall = document.getElementById('filter-hall')?.value || '';
        const status = document.getElementById('filter-status')?.value || '';
        const trainerId = document.getElementById('filter-trainer')?.value || '';
        
        let url = `${API_URL}/trainings`;
        const params = [];
        if (hall) params.push(`hall_type=${hall}`);
        if (status) params.push(`status=${status}`);
        if (trainerId) params.push(`trainer_id=${trainerId}`);
        if (params.length) url += '?' + params.join('&');

        const response = await fetch(url, {
            headers: { 'Authorization': authToken }
        });

        if (!response.ok) throw new Error('Ошибка загрузки');

        const trainings = await response.json();
        
        // Проверяем, что получили массив
        if (!trainings || !Array.isArray(trainings)) {
            console.error('Ожидался массив, получено:', trainings);
            document.getElementById('trainings-list').innerHTML = 
                '<div class="empty-message">Ошибка: неверный формат данных</div>';
            return;
        }
        
        // Сортируем по дате (от ближайших к дальним)
        trainings.sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
        
        displayTrainings(trainings, 'trainings-list');
    } catch (error) {
        console.error('Ошибка:', error);
        document.getElementById('trainings-list').innerHTML = 
            '<div class="empty-message">Ошибка загрузки тренировок</div>';
    }
}

// Отображение тренировок
function displayTrainings(trainings, containerId) {
    const container = document.getElementById(containerId);
    
    if (!container) {
        console.error('Контейнер не найден:', containerId);
        return;
    }
    
    if (!trainings || !Array.isArray(trainings) || trainings.length === 0) {
        container.innerHTML = '<div class="empty-message">Нет доступных тренировок</div>';
        return;
    }

    container.innerHTML = trainings.map(training => {
        const startDate = new Date(training.start_time);
        const endDate = new Date(startDate.getTime() + training.duration_minutes * 60000);
        const isRegistered = training.participants?.some(p => p.user_id === currentUser.id);
        const isTrainer = training.trainer_id === currentUser.id;
        
        // Проверяем, может ли пользователь записаться
        // Админы и тренеры могут записываться без абонемента
        // Обычные пользователи должны иметь активный абонемент (проверка на backend)
        const canRegister = training.status === 'scheduled' && 
                          training.current_participants < training.max_participants &&
                          !isRegistered &&
                          !isTrainer && // Нельзя записаться на свою тренировку как тренер
                          (currentUser.role === 'admin' || currentUser.role === 'trainer' || currentUser.role === 'user'); // Для user проверка абонемента на backend

        const canChangeStatus = currentUser.role === 'admin' || currentUser.role === 'trainer';

        // Определяем роль пользователя в тренировке
        let userRole = '';
        if (isTrainer) {
            userRole = '<span class="training-badge" style="background: #2196F3; color: white;">Вы тренер</span>';
        } else if (isRegistered) {
            userRole = '<span class="training-badge" style="background: #4CAF50; color: white;">Вы участник</span>';
        }

        return `
            <div class="training-card">
                <h3>${training.title}</h3>
                <div class="training-meta">
                    <span class="training-badge badge-hall">${getHallName(training.hall_type)}</span>
                    <span class="training-badge badge-type">${training.type === 'personal' ? 'Персональная' : 'Групповая'}</span>
                    <span class="training-badge badge-status ${training.status}">${getStatusName(training.status)}</span>
                    ${userRole}
                </div>
                <div class="training-info">
                    <p><strong>Тренер:</strong> ${training.trainer?.name || 'N/A'}</p>
                    <p><strong>Время:</strong> ${formatDateTime(startDate)} - ${formatTime(endDate)}</p>
                    <p><strong>Длительность:</strong> ${training.duration_minutes} мин</p>
                    <p><strong>Участников:</strong> ${training.current_participants}/${training.max_participants}</p>
                    ${training.participants && training.participants.length > 0 ? 
                        `<p><strong>Участники:</strong> ${training.participants.map(p => p.user?.name || 'N/A').join(', ')}</p>` : 
                        ''
                    }
                    ${training.description ? `<p>${training.description}</p>` : ''}
                    ${canChangeStatus ? `
                        <div class="status-control">
                            <div class="status-label">Статус тренировки</div>
                            <div class="status-row">
                                <select class="status-select" data-prev-status="${training.status}" onchange="updateTrainingStatus(${training.id}, this)" value="${training.status}">
                                    <option value="scheduled" ${training.status === 'scheduled' ? 'selected' : ''}>Запланировано</option>
                                    <option value="completed" ${training.status === 'completed' ? 'selected' : ''}>Завершено</option>
                                    <option value="cancelled" ${training.status === 'cancelled' ? 'selected' : ''}>Отменено</option>
                                </select>
                            </div>
                        </div>
                    ` : ''}
                </div>
                <div class="training-actions">
                    ${canRegister ? 
                        `<button class="btn btn-primary btn-small" onclick="registerForTraining(${training.id})">Записаться</button>` : 
                        isRegistered ? 
                        `<button class="btn btn-danger btn-small" onclick="cancelRegistration(${training.id})">Отменить запись</button>` : 
                        ''
                    }
                    ${(currentUser.role === 'admin' || isTrainer) ? 
                        `<button class="btn btn-danger btn-small" onclick="deleteTraining(${training.id})">Удалить</button>` : 
                        ''
                    }
                </div>
            </div>
        `;
    }).join('');
}

// Загрузка моих тренировок
async function loadMyTrainings() {
    try {
        const response = await fetch(`${API_URL}/trainings`, {
            headers: { 'Authorization': authToken }
        });
        if (!response.ok) throw new Error('Ошибка загрузки');
        const allTrainings = await response.json();
        
        if (!allTrainings || !Array.isArray(allTrainings)) {
            console.error('Ожидался массив, получено:', allTrainings);
            document.getElementById('my-trainings-list').innerHTML =
                '<div class="empty-message">Ошибка: неверный формат данных</div>';
            return;
        }
        
        // Для всех ролей: показываем тренировки где пользователь тренер ИЛИ участник
        const myTrainings = allTrainings.filter(t => {
            const isTrainer = t.trainer_id === currentUser.id;
            const isParticipant = Array.isArray(t.participants) && 
                                 t.participants.some(p => p.user_id === currentUser.id);
            return isTrainer || isParticipant;
        });
        
        // Сортируем по дате (от ближайших к дальним)
        myTrainings.sort((a, b) => new Date(a.start_time) - new Date(b.start_time));
        
        displayTrainings(myTrainings, 'my-trainings-list');
    } catch (error) {
        console.error('Ошибка:', error);
        document.getElementById('my-trainings-list').innerHTML =
            '<div class="empty-message">Ошибка загрузки тренировок</div>';
    }
}

// Создание тренировки
async function handleCreateTraining(e) {
    e.preventDefault();

    const startTimeInput = document.getElementById('training-start').value;
    if (!startTimeInput) {
        alert('Укажите дату и время тренировки');
        return;
    }

    // Преобразуем локальное время в ISO формат
    const startTime = new Date(startTimeInput);
    const startTimeISO = startTime.toISOString();

    const training = {
        title: document.getElementById('training-title').value,
        description: document.getElementById('training-description').value,
        type: document.getElementById('training-type').value,
        hall_type: document.getElementById('training-hall').value,
        start_time: startTimeISO,
        duration_minutes: parseInt(document.getElementById('training-duration').value),
        max_participants: document.getElementById('training-type').value === 'group' ? 
            parseInt(document.getElementById('training-max').value) : 1,
        trainer_id: parseInt(document.getElementById('training-trainer').value)
    };

    const participantId = document.getElementById('training-participant').value;

    console.log('Создание тренировки:', training);

    try {
        const response = await fetch(`${API_URL}/trainings`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': authToken
            },
            body: JSON.stringify(training)
        });

        if (!response.ok) {
            const error = await response.text();
            console.error('Ошибка создания тренировки:', error);
            alert('Ошибка создания: ' + error);
            return;
        }

        const createdTraining = await response.json();
        console.log('Тренировка создана:', createdTraining);

        // Автоматически регистрируем участника, если выбран (для персональной тренировки)
        if (participantId && training.type === 'personal' && participantId !== '') {
            try {
                const registerResponse = await fetch(`${API_URL}/trainings/${createdTraining.id}/register`, {
                    method: 'POST',
                    headers: {
                        'Authorization': authToken,
                        'X-Participant-Id': participantId.toString() // Отправляем ID участника в заголовке
                    }
                });

                if (!registerResponse.ok) {
                    const errorText = await registerResponse.text();
                    console.warn('Не удалось зарегистрировать участника:', errorText);
                    alert('Тренировка создана, но не удалось зарегистрировать участника: ' + errorText);
                } else {
                    console.log('Участник успешно зарегистрирован');
                }
            } catch (regError) {
                console.error('Ошибка регистрации участника:', regError);
                alert('Тренировка создана, но произошла ошибка при регистрации участника');
            }
        }

        closeTrainingModal();
        loadTrainings();
        alert('Тренировка успешно создана!');
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка создания тренировки: ' + error.message);
    }
}

// Регистрация на тренировку
async function registerForTraining(trainingId) {
    if (!confirm('Записаться на эту тренировку?')) return;

    try {
        const response = await fetch(`${API_URL}/trainings/${trainingId}/register`, {
            method: 'POST',
            headers: { 'Authorization': authToken }
        });

        if (!response.ok) {
            const error = await response.text();
            alert('Ошибка: ' + error);
            return;
        }

        loadTrainings();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка регистрации');
    }
}

// Отмена регистрации
async function cancelRegistration(trainingId) {
    if (!confirm('Отменить запись на эту тренировку?')) return;

    try {
        const response = await fetch(`${API_URL}/trainings/${trainingId}/cancel`, {
            method: 'POST',
            headers: { 'Authorization': authToken }
        });

        if (!response.ok) {
            const error = await response.text();
            alert('Ошибка: ' + error);
            return;
        }

        loadTrainings();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка отмены регистрации');
    }
}

// Удаление тренировки
async function deleteTraining(trainingId) {
    if (!confirm('Удалить эту тренировку?')) return;

    try {
        const response = await fetch(`${API_URL}/trainings/${trainingId}`, {
            method: 'DELETE',
            headers: { 'Authorization': authToken }
        });

        if (!response.ok) {
            const error = await response.text();
            alert('Ошибка: ' + error);
            return;
        }

        loadTrainings();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка удаления');
    }
}

// Модальное окно тренировки
function closeTrainingModal() {
    document.getElementById('training-modal').classList.remove('active');
    document.getElementById('training-form').reset();
}

async function openTrainingModal() {
    const modal = document.getElementById('training-modal');
    modal.classList.add('active');
    
    // Сбрасываем значения по умолчанию
    const form = document.getElementById('training-form');
    if (form) form.reset();
    const typeSelect = document.getElementById('training-type');
    if (typeSelect) {
        typeSelect.value = 'personal'; // всегда персональная по умолчанию
        toggleMaxParticipants();
    }
    
    // Загружаем список тренеров
    await loadTrainersForSelect();
    
    // Загружаем список пользователей
    await loadUsersForSelect();
}

async function loadTrainersForSelect() {
    try {
        const response = await fetch(`${API_URL}/users`, {
            headers: { 'Authorization': authToken }
        });
        const users = await response.json();
        const select = document.getElementById('training-trainer');
        select.innerHTML = '<option value="">Выберите тренера</option>';
        
        const trainers = users.filter(u => u.role === 'trainer' || u.role === 'admin');
        trainers.forEach(trainer => {
            const option = document.createElement('option');
            option.value = trainer.id;
            option.textContent = `${trainer.name} (${trainer.email})`;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('Ошибка загрузки тренеров:', error);
    }
}

async function loadUsersForSelect() {
    try {
        const response = await fetch(`${API_URL}/users`, {
            headers: { 'Authorization': authToken }
        });
        const users = await response.json();
        const select = document.getElementById('training-participant');
        select.innerHTML = '<option value="">Не выбран</option>';
        
        // Показываем всех пользователей (любая роль может быть участником)
        users.forEach(user => {
            const option = document.createElement('option');
            option.value = user.id;
            option.textContent = `${user.name} (${user.email}) - ${user.role}`;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('Ошибка загрузки пользователей:', error);
    }
}

function toggleMaxParticipants() {
    const type = document.getElementById('training-type').value;
    const group = document.getElementById('max-participants-group');
    const participantSelect = document.getElementById('training-participant').parentElement;
    
    if (type === 'group') {
        group.style.display = 'block';
        participantSelect.style.display = 'none';
    } else {
        group.style.display = 'none';
        participantSelect.style.display = 'block';
    }
}

// Вспомогательные функции
function getHallName(hall) {
    const names = {
        'pilates': 'Пилатес',
        'yoga': 'Йога',
        'gym': 'Тренажерный зал',
        'dance': 'Танцевальный зал',
        'cardio': 'Кардио зал'
    };
    return names[hall] || hall;
}

function getStatusName(status) {
    const names = {
        'scheduled': 'Запланировано',
        'completed': 'Завершено',
        'cancelled': 'Отменено'
    };
    return names[status] || status;
}

function formatDateTime(date) {
    return date.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatTime(date) {
    return date.toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Функции для админ-панели
async function loadUsers() {
    try {
        const response = await fetch(`${API_URL}/users`, {
            headers: { 'Authorization': authToken }
        });
        
        if (!response.ok) {
            throw new Error('Ошибка загрузки пользователей');
        }
        
        const users = await response.json();
        const list = document.getElementById('users-list');
        
        if (!list) {
            console.error('Элемент users-list не найден');
            return;
        }
        
        if (!users || !Array.isArray(users)) {
            list.innerHTML = '<div class="empty-message">Ошибка: неверный формат данных</div>';
            return;
        }
        
        list.innerHTML = '';
        
        if (users.length === 0) {
            list.innerHTML = '<div class="empty-message">Нет пользователей</div>';
        } else {
            users.forEach(u => {
                list.innerHTML += `
                  <div class="list-item">
                    <div class="list-item-info">
                      <p><strong>${u.name}</strong> (${u.email})</p>
                      <p>Роль: ${u.role}</p>
                    </div>
                    ${currentUser && currentUser.role === 'admin' ? `
                    <div style="display: flex; gap: 10px;">
                      <button class="btn btn-danger btn-small" onclick="deleteUser(${u.id})">Удалить</button>
                    </div>
                    ` : ''}
                  </div>
                `;
            });
        }
        
        // Кнопка добавления всегда внизу
        list.innerHTML += `<div style="margin-top:20px;"><button class="btn btn-primary" onclick="showUserModal(null)">+ Добавить пользователя</button></div>`;
    } catch (error) {
        console.error('Ошибка загрузки пользователей:', error);
        const list = document.getElementById('users-list');
        if (list) {
            list.innerHTML = '<div class="empty-message">Ошибка загрузки пользователей</div>';
        }
    }
}

async function deleteUser(id) {
    if (!confirm('Удалить пользователя?')) return;
    try {
        await fetch(`${API_URL}/users/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': authToken }
        });
        loadUsers();
    } catch (error) {
        console.error('Ошибка:', error);
    }
}

async function loadClients() {
    try {
        const response = await fetch(`${API_URL}/clients`, {
            headers: { 'Authorization': authToken }
        });
        const clients = await response.json();
        const list = document.getElementById('clients-list');
        
        if (!clients || !Array.isArray(clients) || clients.length === 0) {
            list.innerHTML = '<div class="empty-message">Нет клиентов</div>';
            return;
        }
        
        // Загружаем абонементы для каждого клиента
        const subscriptionsResponse = await fetch(`${API_URL}/subscriptions`, {
            headers: { 'Authorization': authToken }
        });
        const allSubscriptions = await subscriptionsResponse.json();
        
        list.innerHTML = clients.map(c => {
            // Находим активные абонементы для этого клиента
            const activeSubscriptions = (allSubscriptions || []).filter(s => 
                s.client_id === c.id && 
                s.status === 'active' && 
                new Date(s.end_date) >= new Date()
            );
            
            const hasActiveSubscription = activeSubscriptions.length > 0;
            
            return `
            <div class="list-item">
                <div class="list-item-info">
                    <p><strong>${c.user?.name || 'N/A'}</strong> (${c.user?.email || 'N/A'})</p>
                    <p>Телефон: ${c.phone || 'не указан'}</p>
                    <p>Адрес: ${c.address || 'не указан'}</p>
                    <p>Абонемент: ${hasActiveSubscription ? 
                        `<span class="badge badge-status scheduled">Активен (${activeSubscriptions[0].type})</span>` : 
                        '<span class="badge badge-status cancelled">Нет активного абонемента</span>'
                    }</p>
                </div>
                ${currentUser && currentUser.role === 'admin' ? `
                <button class="btn btn-danger btn-small" onclick="deleteClient(${c.id})">Удалить</button>
                ` : ''}
            </div>
        `;
        }).join('');
    } catch (error) {
        console.error('Ошибка:', error);
        document.getElementById('clients-list').innerHTML = '<div class="empty-message">Ошибка загрузки клиентов</div>';
    }
}

async function deleteClient(id) {
    if (!confirm('Удалить этого клиента?')) return;
    try {
        const response = await fetch(`${API_URL}/clients/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': authToken }
        });
        if (!response.ok) {
            const error = await response.text();
            alert('Ошибка: ' + error);
            return;
        }
        loadClients();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка удаления клиента');
    }
}

async function loadSubscriptions() {
    try {
        const response = await fetch(`${API_URL}/subscriptions`, {
            headers: { 'Authorization': authToken }
        });
        const subscriptions = await response.json();
        const list = document.getElementById('subscriptions-list');
        
        if (!subscriptions || !Array.isArray(subscriptions) || subscriptions.length === 0) {
            list.innerHTML = '<div class="empty-message">Нет абонементов</div>';
            return;
        }
        
        list.innerHTML = subscriptions.map(s => {
            const startDate = s.start_date ? new Date(s.start_date).toLocaleDateString('ru-RU') : 'N/A';
            const endDate = s.end_date ? new Date(s.end_date).toLocaleDateString('ru-RU') : 'N/A';
            const clientName = s.client?.user?.name || 'N/A';
            return `
            <div class="list-item">
                <div class="list-item-info">
                    <p><strong>${s.type}</strong> - ${clientName}</p>
                    <p>Период: ${startDate} - ${endDate}</p>
                    <p>Цена: ${s.price} руб.</p>
                    <p>Статус: <span class="badge ${s.status === 'active' ? 'badge-status scheduled' : 'badge-status cancelled'}">${s.status === 'active' ? 'Активен' : s.status === 'expired' ? 'Истек' : 'Отменен'}</span></p>
                </div>
                ${currentUser && currentUser.role === 'admin' ? `
                <button class="btn btn-danger btn-small" onclick="deleteSubscription(${s.id})">Удалить</button>
                ` : ''}
            </div>
        `;
        }).join('');
    } catch (error) {
        console.error('Ошибка:', error);
        document.getElementById('subscriptions-list').innerHTML = '<div class="empty-message">Ошибка загрузки абонементов</div>';
    }
}

async function deleteSubscription(id) {
    if (!confirm('Удалить этот абонемент?')) return;
    try {
        const response = await fetch(`${API_URL}/subscriptions/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': authToken }
        });
        if (!response.ok) {
            const error = await response.text();
            alert('Ошибка: ' + error);
            return;
        }
        loadSubscriptions();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Ошибка удаления абонемента');
    }
}

// Модальное окно создания абонемента
async function showSubscriptionModal() {
    // Загружаем пользователей с ролью "user"
    let users = [];
    try {
        const response = await fetch(`${API_URL}/users`, {
            headers: { 'Authorization': authToken }
        });
        if (!response.ok) {
            const errorText = await response.text();
            console.error('Ошибка загрузки пользователей:', response.status, errorText);
            throw new Error(`Ошибка загрузки пользователей: ${response.status}`);
        }
        const data = await response.json();
        if (Array.isArray(data)) {
            users = data.filter(u => u.role === 'user');
        } else {
            console.error('Ожидался массив пользователей, получено:', typeof data, data);
        }
    } catch (error) {
        console.error('Ошибка загрузки пользователей:', error);
    }
    
    const modal = document.createElement('div');
    modal.id = 'subscription-modal';
    modal.className = 'modal active';
    modal.innerHTML = `
        <div class='modal-content'>
            <span class='close' onclick='closeSubscriptionModal()'>&times;</span>
            <h2>Создать абонемент</h2>
            <form id='subscription-form'>
                <div class='form-group'>
                    <label>Пользователь</label>
                    <select id='sub-user-id' required>
                        <option value=''>Выберите пользователя</option>
                    </select>
                    <small style="color: #666;">Выберите пользователя с ролью "user"</small>
                </div>
                <div class='form-group'>
                    <label>Тип абонемента</label>
                    <select id='sub-type' required>
                        <option value=''>Выберите тип</option>
                        <option value='monthly'>Месячный (2000 руб.)</option>
                        <option value='quarterly'>Квартальный (5000 руб.)</option>
                        <option value='yearly'>Годовой (18000 руб.)</option>
                    </select>
                </div>
                <div class='form-group'>
                    <label>Дата начала</label>
                    <input id='sub-start-date' type='date' required>
                    <small style="color: #666;">Дата окончания и статус будут рассчитаны автоматически</small>
                </div>
                <div class='form-actions'>
                    <button type='button' class='btn btn-secondary' onclick='closeSubscriptionModal()'>Отмена</button>
                    <button type='submit' class='btn btn-primary'>Создать</button>
                </div>
            </form>
        </div>`;
    
    document.body.appendChild(modal);
    
    // Заполняем список пользователей
    const userSelect = document.getElementById('sub-user-id');
    if (users.length === 0) {
        userSelect.innerHTML = '<option value=\"\">Нет доступных пользователей</option>';
        userSelect.disabled = true;
    } else {
        users.forEach(user => {
            const option = document.createElement('option');
            option.value = user.id;
            option.textContent = `${user.name} (${user.email})`;
            userSelect.appendChild(option);
        });
        userSelect.disabled = false;
    }
    
    // Устанавливаем сегодняшнюю дату по умолчанию
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('sub-start-date').value = today;
    
    // Обработчик формы
    document.getElementById('subscription-form').onsubmit = async function(e) {
        e.preventDefault();
        
        const userId = document.getElementById('sub-user-id').value;
        if (!userId) {
            alert('Пожалуйста, выберите пользователя');
            return;
        }
        
        const typeValue = document.getElementById('sub-type').value;
        if (!typeValue) {
            alert('Пожалуйста, выберите тип абонемента');
            return;
        }
        
        const startDateValue = document.getElementById('sub-start-date').value;
        if (!startDateValue) {
            alert('Пожалуйста, укажите дату начала');
            return;
        }
        
        // Формируем данные для отправки (только user_id, type и start_date)
        const data = {
            user_id: parseInt(userId),
            type: typeValue,
            start_date: startDateValue // Формат: YYYY-MM-DD
        };
        
        console.log('Отправка данных абонемента:', data);
        
        try {
            const response = await fetch(`${API_URL}/subscriptions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': authToken
                },
                body: JSON.stringify(data)
            });
            
            if (!response.ok) {
                const error = await response.text();
                console.error('Ошибка создания абонемента:', error);
                alert('Ошибка: ' + error);
                return;
            }
            
            closeSubscriptionModal();
            loadSubscriptions();
            // Обновляем список клиентов, если открыта вкладка
            const activeTab = document.querySelector('.tab-content.active');
            if (activeTab && activeTab.id === 'clients-tab') {
                loadClients();
            }
            alert('Абонемент успешно создан!');
        } catch (error) {
            console.error('Ошибка создания абонемента:', error);
            alert('Ошибка: ' + error.message);
        }
    };
}

function closeSubscriptionModal() {
    const modal = document.getElementById('subscription-modal');
    if (modal) {
        modal.remove();
    }
}

async function loadEmployees() {
    try {
        const response = await fetch(`${API_URL}/employees`, {
            headers: { 'Authorization': authToken }
        });
        const employees = await response.json();
        const list = document.getElementById('employees-list');
        list.innerHTML = employees.map(e => `
            <div class="list-item">
                <div class="list-item-info">
                    <p><strong>${e.user?.name || 'N/A'}</strong></p>
                    <p>Должность: ${e.position}</p>
                    <p>Зарплата: ${e.salary || 'не указана'}</p>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Ошибка:', error);
    }
}

// Модальное окно для создания/редактирования пользователя
function showUserModal(userId) {
    console.log('showUserModal вызвана с userId:', userId);
    
    // Удаляем существующее модальное окно, если есть
    const existingModal = document.getElementById('user-modal');
    if (existingModal) {
        existingModal.remove();
    }
    
    // Создаем новое модальное окно
    const modal = document.createElement('div');
    modal.id = 'user-modal';
    modal.className = 'modal active';
    modal.innerHTML = `
        <div class='modal-content'>
            <span class='close' onclick='closeUserModal()'>&times;</span>
            <h2>${userId ? 'Редактировать пользователя' : 'Создать пользователя'}</h2>
                <form id='user-form-modal'>
                    <div class='form-group'>
                        <label>Имя</label>
                        <input id='um-name' type='text' required>
                    </div>
                    <div class='form-group'>
                        <label>Email</label>
                        <input id='um-email' type='email' required>
                    </div>
                    <div class='form-group'>
                        <label>Пароль</label>
                        <input id='um-password' type='password' placeholder="${userId ? 'Оставьте пустым, чтобы не менять' : 'Оставьте пустым для автогенерации'}" autocomplete="new-password">
                        <small style="color: #666; font-size: 12px;">${userId ? '' : 'Если не указан, будет сгенерирован случайный пароль'}</small>
                    </div>
                    <div class='form-group'>
                        <label>Роль</label>
                        <select id='um-role' required>
                            <option value='user'>Пользователь (автоматически создастся клиент)</option>
                            <option value='trainer'>Тренер (автоматически создастся сотрудник)</option>
                            <option value='admin'>Администратор</option>
                        </select>
                    </div>
                    <div class='form-actions'>
                        <button type='button' class='btn btn-secondary' onclick='closeUserModal()'>Отмена</button>
                        <button type='submit' class='btn btn-primary'>${userId ? 'Сохранить' : 'Создать'}</button>
                    </div>
                </form>
        </div>`;
    
    document.body.appendChild(modal);
    
    // Обработчик отправки формы
    document.getElementById('user-form-modal').onsubmit = async function(e) {
        e.preventDefault();
        
        const data = {
            name: document.getElementById('um-name').value.trim(),
            email: document.getElementById('um-email').value.trim(),
            role: document.getElementById('um-role').value
        };
        
        // Пароль опциональный - если не указан, сервер сгенерирует случайный
        const password = document.getElementById('um-password').value.trim();
        if (password) {
            data.password = password;
        }
        
        console.log('Создание пользователя:', { ...data, password: password ? '***' : '(автогенерация)' });
        
        try {
            const resp = await fetch(`${API_URL}/users`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': authToken
                },
                body: JSON.stringify(data)
            });
            
            if (!resp.ok) {
                const errorText = await resp.text();
                alert('Ошибка: ' + errorText);
                return;
            }
            
            const createdUser = await resp.json();
            let message = 'Пользователь успешно создан!';
            
            // Информируем о создании связанных записей
            if (data.role === 'user') {
                message += '\nАвтоматически создан клиент.';
            } else if (data.role === 'trainer') {
                message += '\nАвтоматически создан сотрудник.';
            }
            
            if (!password) {
                message += '\n\nВНИМАНИЕ: Пароль был сгенерирован автоматически.';
                message += '\nПопросите пользователя использовать функцию восстановления пароля или установите пароль вручную.';
            }
            
            closeUserModal();
            loadUsers();
            
            // Обновляем списки клиентов и сотрудников, если открыты соответствующие вкладки
            const activeTab = document.querySelector('.tab-content.active');
            if (activeTab && activeTab.id === 'clients-tab') {
                loadClients();
            } else if (activeTab && activeTab.id === 'employees-tab') {
                loadEmployees();
            }
            
            alert(message);
        } catch (error) {
            console.error('Ошибка создания пользователя:', error);
            alert('Ошибка: ' + error.message);
        }
    };
}

function closeUserModal() {
    const modal = document.getElementById('user-modal');
    if (modal) {
        modal.remove();
    }
}
