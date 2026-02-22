import sys
import math
import random
import json
import csv
from datetime import datetime, timedelta
from PyQt5.QtWidgets import (QApplication, QMainWindow, QVBoxLayout, QWidget, 
                            QSlider, QLabel, QHBoxLayout, QGridLayout,
                            QPushButton, QCheckBox, QListWidget, QListWidgetItem,
                            QGroupBox)
from PyQt5.QtWebEngineWidgets import QWebEngineView
from PyQt5.QtCore import Qt, QTimer, pyqtSignal, QThread
from PyQt5.QtGui import QFont

class WorkerThread(QThread):
    """Поток для тяжелых вычислений"""
    result_ready = pyqtSignal(dict)
    
    def __init__(self, agents, location_coords, current_time, base_time, selected_agents):
        super().__init__()
        self.agents = agents
        self.location_coords = location_coords
        self.current_time = current_time
        self.base_time = base_time
        self.selected_agents = selected_agents
        self.total_seconds = 0
    
    def run(self):
        """Основной метод потока"""
        coords = {}
        for agent_id in self.selected_agents:
            if agent_id in self.agents:
                coords_data = self.get_agent_coords_at_time(agent_id)
                if coords_data:
                    coords[agent_id] = coords_data
        
        self.result_ready.emit(coords)
    
    def get_agent_coords_at_time(self, agent_id):
        """
        Получает координаты агента в заданное время с использованием интерполяции
        
        Алгоритм:
        1. Преобразуем абсолютное время в часы относительно базового времени
        2. Сортируем точки маршрута агента по времени
        3. Ищем интервал между двумя точками маршрута, в который попадает текущее время
        4. Определяем, находимся ли мы в "окне движения" (20 минут вокруг запланированного времени перехода)
        5. Если в окне движения - интерполируем положение между точками
        6. Если вне окна - остаемся на предыдущей точке
        
        Возвращает (lat, lon) или None, если данные недоступны
        """
        if agent_id not in self.agents:
            return None
        
        agent_data = self.agents[agent_id]
        if not agent_data['locations']:
            return None
        
        # Шаг 1: Преобразуем абсолютное время в Unix timestamp в часы с дробной частью
        # Пример: base_time = 1704067200 (2024-01-01 00:00:00 UTC)
        # current_time = 1704070800 (2024-01-01 01:00:00 UTC)
        # hour = 1.0
        hour = (self.current_time - self.base_time) / 3600.0
        
        # Шаг 2: Сортируем точки маршрута по времени
        # locations = [(час1, 'место1'), (час2, 'место2'), ...]
        locations = sorted(agent_data['locations'], key=lambda x: x[0])
        
        # Шаг 3: Ищем интервал, в котором находится текущее время
        for i in range(len(locations) - 1):
            hour1, loc1 = locations[i]    # Начало интервала: время и место
            hour2, loc2 = locations[i + 1]  # Конец интервала: время и место
            
            # Проверяем, попадает ли текущее время в этот интервал (до часа hour2)
            if hour <= hour2:
                # Шаг 4: Определяем 20-минутное окно движения
                # Движение происходит в течение 20 минут: 10 мин до и 10 мин после часа hour2
                movement_start = hour2 - 10/60  # 10 минут до часа hour2 (в часах)
                movement_end = hour2 + 10/60    # 10 минут после часа hour2 (в часах)
                
                # Получаем координаты начальной и конечной точек
                coords1 = self.location_coords.get(loc1)
                coords2 = self.location_coords.get(loc2)
                
                # Если начальные координаты не найдены, данные некорректны
                if not coords1:
                    return None
                
                # Шаг 5: Проверяем, находимся ли мы в окне движения
                if movement_start <= hour <= movement_end:
                    # Мы в окне движения - выполняем линейную интерполяцию
                    
                    # Если конечные координаты не найдены, остаемся на начальной точке
                    if not coords2:
                        return coords1
                    
                    # Вычисляем коэффициент интерполяции t от 0 до 1
                    # t = 0 в начале движения (movement_start)
                    # t = 1 в конце движения (movement_end)
                    t = (hour - movement_start) / (movement_end - movement_start)
                    t = max(0, min(1, t))  # Ограничиваем t в диапазоне [0, 1]
                    
                    # Линейная интерполяция координат:
                    # Новая координата = начальная + t * (конечная - начальная)
                    # Формула работает для каждой оси (широты и долготы)
                    lat = coords1[0] + t * (coords2[0] - coords1[0])
                    lon = coords1[1] + t * (coords2[1] - coords1[1])
                    
                    return (lat, lon)
                else:
                    # Шаг 6: Мы не в окне движения - агент остается на текущей точке
                    # (в начальной точке интервала до начала движения)
                    return coords1
        
        # Если прошли все интервалы и не нашли подходящий,
        # значит текущее время больше всех запланированных времен
        # Возвращаем координаты последней известной точки
        return self.location_coords.get(locations[-1][1])

class MapWindow(QMainWindow):
    update_marker_signal = pyqtSignal(list, str)
    
    def __init__(self):
        super().__init__()
        self.setWindowTitle("Траектория движения объектов")
        self.setGeometry(100, 100, 1200, 900)
        
        # Инициализируем переменные
        self.agents = {}
        self.location_coords = {}
        self.is_page_loaded = False
        self.agent_colors = {}
        self.selected_agents = set()
        self.show_trajectories = False
        self.worker_thread = None
        
        # Кэш для координат (оптимизация)
        self.coord_cache = {}
        self.cache_time = 0
        
        # Создаем центральный виджет
        central_widget = QWidget()
        self.setCentralWidget(central_widget)
        main_layout = QVBoxLayout(central_widget)
        
        # Верхняя панель с информацией
        info_layout = QHBoxLayout()
        self.time_label = QLabel("Время: не установлено")
        self.time_label.setFont(QFont("Arial", 10))
        self.coord_label = QLabel("Агентов отображается: 0")
        self.coord_label.setFont(QFont("Arial", 10))
        self.memory_label = QLabel("Память: -")
        self.memory_label.setFont(QFont("Arial", 10))
        
        info_layout.addWidget(self.time_label)
        info_layout.addWidget(self.coord_label)
        info_layout.addWidget(self.memory_label)
        info_layout.addStretch()
        main_layout.addLayout(info_layout)
        
        # Основной контейнер с картой и списком агентов
        content_layout = QHBoxLayout()
        
        # Создаем WebEngineView для отображения карты
        self.web_view = QWebEngineView()
        content_layout.addWidget(self.web_view, 3)
        
        # Правая панель с элементами управления
        control_panel = QWidget()
        control_panel.setMaximumWidth(300)
        control_layout = QVBoxLayout(control_panel)
        
        # Группа выбора агентов
        agents_group = QGroupBox("Выбор агентов")
        agents_layout = QVBoxLayout()
        
        # Кнопки управления выбором
        buttons_layout = QHBoxLayout()
        self.select_all_btn = QPushButton("Выбрать все")
        self.select_all_btn.clicked.connect(self.select_all_agents)
        self.clear_all_btn = QPushButton("Очистить все")
        self.clear_all_btn.clicked.connect(self.clear_all_agents)
        
        buttons_layout.addWidget(self.select_all_btn)
        buttons_layout.addWidget(self.clear_all_btn)
        agents_layout.addLayout(buttons_layout)
        
        # Список агентов с чекбоксами
        self.agents_list = QListWidget()
        agents_layout.addWidget(self.agents_list)
        
        agents_group.setLayout(agents_layout)
        control_layout.addWidget(agents_group)
        
        # Чекбокс для траекторий
        self.trajectory_checkbox = QCheckBox("Показать траектории")
        self.trajectory_checkbox.setChecked(False)
        self.trajectory_checkbox.stateChanged.connect(self.on_trajectory_changed)
        control_layout.addWidget(self.trajectory_checkbox)
        
        # Группа управления временем
        time_group = QGroupBox("Управление временем")
        time_layout = QVBoxLayout()
        
        # Метки времени
        time_info_layout = QHBoxLayout()
        self.current_time_label = QLabel("00:00")
        self.current_time_label.setFont(QFont("Arial", 12, QFont.Bold))
        self.current_date_label = QLabel("01.01.2024")
        self.current_date_label.setFont(QFont("Arial", 10))
        
        time_info_layout.addWidget(self.current_time_label)
        time_info_layout.addWidget(self.current_date_label)
        time_layout.addLayout(time_info_layout)
        
        # Слайдер времени
        self.time_slider = QSlider(Qt.Horizontal)
        self.time_slider.setMinimum(0)
        self.time_slider.setMaximum(1000)
        self.time_slider.valueChanged.connect(self.on_slider_changed)
        time_layout.addWidget(self.time_slider)
        
        # Кнопки управления воспроизведением
        play_buttons_layout = QHBoxLayout()
        self.prev_button = QPushButton("◀◀")
        self.prev_button.clicked.connect(self.prev_frame_step)
        self.play_button = QPushButton("▶")
        self.play_button.clicked.connect(self.toggle_play)
        self.next_button = QPushButton("▶▶")
        self.next_button.clicked.connect(self.next_frame_step)
        
        play_buttons_layout.addWidget(self.prev_button)
        play_buttons_layout.addWidget(self.play_button)
        play_buttons_layout.addWidget(self.next_button)
        time_layout.addLayout(play_buttons_layout)
        
        # Кнопка для быстрого перехода к началу
        self.reset_button = QPushButton("В начало")
        self.reset_button.clicked.connect(self.reset_to_start)
        time_layout.addWidget(self.reset_button)
        
        time_group.setLayout(time_layout)
        control_layout.addWidget(time_group)
        
        # Скорость воспроизведения
        speed_group = QGroupBox("Скорость воспроизведения")
        speed_layout = QVBoxLayout()
        
        self.speed_slider = QSlider(Qt.Horizontal)
        self.speed_slider.setMinimum(1)
        self.speed_slider.setMaximum(10)
        self.speed_slider.setValue(3)
        self.speed_slider.valueChanged.connect(self.on_speed_changed)
        self.speed_label = QLabel("Скорость: 1x")
        speed_layout.addWidget(self.speed_label)
        speed_layout.addWidget(self.speed_slider)
        
        speed_group.setLayout(speed_layout)
        control_layout.addWidget(speed_group)
        
        control_layout.addStretch()
        content_layout.addWidget(control_panel)
        
        main_layout.addLayout(content_layout)
        
        # Флаг воспроизведения
        self.is_playing = False
        self.play_timer = QTimer()
        self.play_timer.timeout.connect(self.next_frame)
        self.play_speed = 1.0  # множитель скорости по умолчанию
        
        # Загружаем данные из файла
        self.load_data_from_file("demo.csv")
        
        # Генерируем координаты для мест
        self.generate_location_coordinates()
        
        # Подключаем сигнал для обновления маркеров
        self.update_marker_signal.connect(self.update_marker_positions)
        
        # Ждем загрузки страницы
        self.web_view.loadFinished.connect(self.on_page_loaded)
        
        # Загружаем карту
        self.load_osm_map()
        
        # Таймер для обновления информации о памяти
        self.memory_timer = QTimer()
        self.memory_timer.timeout.connect(self.update_memory_info)
        self.memory_timer.start(2000)
    
    def update_memory_info(self):
        """Обновляет информацию об использовании памяти"""
        try:
            import psutil
            process = psutil.Process()
            memory_mb = process.memory_info().rss / 1024 / 1024
            self.memory_label.setText(f"Память: {memory_mb:.1f} MB")
        except:
            self.memory_label.setText("Память: N/A")
    
    def generate_location_coordinates(self):
        """Генерирует случайные координаты только для мест без координат из CSV"""
        # Границы Москвы (примерные)
        moscow_bounds = {
            'lat_min': 55.55, 'lat_max': 55.91,
            'lon_min': 37.37, 'lon_max': 37.84
        }
        
        # Собираем все уникальные места
        unique_locations = set()
        for agent_data in self.agents.values():
            for hour, location in agent_data['locations']:
                unique_locations.add(location)
        
        # Генерируем координаты только для мест, которые не имеют координат из CSV
        generated_count = 0
        for location in unique_locations:
            if location not in self.location_coords:
                lat = random.uniform(moscow_bounds['lat_min'], moscow_bounds['lat_max'])
                lon = random.uniform(moscow_bounds['lon_min'], moscow_bounds['lon_max'])
                self.location_coords[location] = (lat, lon)
                generated_count += 1
        
        print(f"Загружено координат из CSV: {len(self.location_coords) - generated_count}")
        print(f"Сгенерировано случайных координат: {generated_count}")
        print(f"Всего мест с координатами: {len(self.location_coords)}")
    
    def load_data_from_file(self, filename):
        """Загрузка данных из CSV файла с оптимизацией"""
        self.agents = {}
        try:
            with open(filename, 'r', encoding='utf-8') as f:
                csv_reader = csv.reader(f)
                header = next(csv_reader)  # Читаем заголовок
                
                print(f"Загрузка данных из {filename}...")
                print(f"Заголовки CSV: {header}")
                
                # Создаем словарь для хранения последнего местоположения каждого агента
                last_locations = {}
                
                for line_num, row in enumerate(csv_reader, 1):
                    if len(row) < 11:  # Проверяем количество колонок
                        continue
                    
                    try:
                        hour = int(row[0])
                        agent_id = int(row[1])
                        location = row[6]
                        geo_coords = row[10]  # Координаты из колонки geo
                        
                        # Парсим координаты из строки "lat,lon"
                        if geo_coords and geo_coords.strip() and geo_coords.strip() != " ":
                            try:
                                lat_str, lon_str = geo_coords.split(',')
                                lat, lon = float(lat_str), float(lon_str)
                                # Сохраняем координаты для этого места
                                self.location_coords[location] = (lat, lon)
                            except (ValueError, IndexError):
                                pass  # Игнорируем некорректные координаты
                        
                        # Инициализируем агента если его еще нет
                        if agent_id not in self.agents:
                            self.agents[agent_id] = {
                                'locations': [],
                                'last_hour': -1,
                                'last_location': None
                            }
                        
                        # Сохраняем только если местоположение изменилось
                        current_location = location
                        if (agent_id not in last_locations or
                            last_locations[agent_id] != current_location or
                            hour - self.agents[agent_id]['last_hour'] > 24):
                            
                            self.agents[agent_id]['locations'].append((hour, location))
                            self.agents[agent_id]['last_hour'] = hour
                            self.agents[agent_id]['last_location'] = location
                            last_locations[agent_id] = current_location
                    
                    except (ValueError, IndexError):
                        continue
                    
                    # Прогресс каждые 10000 строк
                    if line_num % 10000 == 0:
                        print(f"Обработано {line_num} строк...")
            
            print(f"Загружено {len(self.agents)} агентов")
            
            # Генерируем цвета для агентов
            colors = ['#FF0000', '#00FF00', '#0000FF', '#FFFF00', '#FF00FF', 
                     '#00FFFF', '#FFA500', '#800080', '#FF69B4', '#32CD32']
            for i, agent_id in enumerate(self.agents.keys()):
                self.agent_colors[agent_id] = colors[i % len(colors)]
            
            # По умолчанию выбираем первых 5 агентов для производительности
            self.selected_agents = set(list(self.agents.keys())[:5])
            
            # Заполняем список агентов
            self.update_agents_list()
            
            # Инициализируем данные
            self.init_data()
            
            return True
            
        except FileNotFoundError:
            print(f"Файл {filename} не найден.")
            return False
        except Exception as e:
            print(f"Ошибка при загрузке файла: {e}")
            return False
    
    def update_agents_list(self):
        """Обновляет список агентов с чекбоксами"""
        self.agents_list.clear()
        for agent_id in sorted(self.agents.keys()):
            item = QListWidgetItem(f"Агент {agent_id}")
            item.setData(Qt.UserRole, agent_id)
            item.setFlags(item.flags() | Qt.ItemIsUserCheckable)
            item.setCheckState(Qt.Checked if agent_id in self.selected_agents else Qt.Unchecked)
            self.agents_list.addItem(item)
        
        # Подключаем сигнал изменения состояния чекбоксов
        self.agents_list.itemChanged.connect(self.on_agent_selection_changed)
    
    def on_agent_selection_changed(self, item):
        """Обработчик изменения выбора агента"""
        agent_id = item.data(Qt.UserRole)
        if item.checkState() == Qt.Checked:
            self.selected_agents.add(agent_id)
        else:
            self.selected_agents.discard(agent_id)
        
        print(f"Выбрано {len(self.selected_agents)} агентов")
        # Очищаем кэш при изменении выбора агентов
        self.coord_cache.clear()
        self.update_display()
    
    def select_all_agents(self):
        """Выбирает всех агентов"""
        for i in range(self.agents_list.count()):
            item = self.agents_list.item(i)
            item.setCheckState(Qt.Checked)
    
    def clear_all_agents(self):
        """Очищает выбор всех агентов"""
        for i in range(self.agents_list.count()):
            item = self.agents_list.item(i)
            item.setCheckState(Qt.Unchecked)
    
    def on_trajectory_changed(self, state):
        """Обработчик изменения состояния чекбокса траекторий"""
        self.show_trajectories = (state == Qt.Checked)
        if self.is_page_loaded:
            self.update_trajectories()
    
    def on_speed_changed(self, value):
        """Обработчик изменения скорости воспроизведения"""
        # Преобразуем значение слайдера в множитель скорости
        speed_multipliers = [0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 10.0, 20.0]
        self.play_speed = speed_multipliers[value - 1]
        self.speed_label.setText(f"Скорость: {self.play_speed}x")
    
    def init_data(self):
        """Инициализация данных после загрузки"""
        if not self.agents:
            return
        
        # Минимальное и максимальное время (в часах)
        self.min_hour = 0
        self.max_hour = 0
        for agent_id, agent_data in self.agents.items():
            if agent_data['locations']:
                max_hour = max(h for h, _ in agent_data['locations'])
                if max_hour > self.max_hour:
                    self.max_hour = max_hour
        
        # Преобразуем часы в timestamp (начинаем с 2024-01-01)
        self.base_time = datetime(2024, 1, 1).timestamp()
        self.min_time = self.base_time
        self.max_time = self.base_time + (self.max_hour * 3600)
        
        # Устанавливаем начальное положение
        self.current_time = self.min_time
        self.current_seconds = 0
        self.total_seconds = self.max_hour * 3600  # Общее количество секунд
        self.frame_step = 30  # Шаг в секундах для плавного движения
        
        # Обновляем отображение
        self.update_display()
    
    def seconds_to_time(self, seconds):
        """Конвертация секунд во время"""
        return self.base_time + seconds
    
    def time_to_seconds(self, time):
        """Конвертация времени в секунды"""
        return int(time - self.base_time)
    
    def time_to_slider_value(self, time):
        """Конвертация времени в значение слайдера"""
        if self.total_seconds == 0:
            return 0
        current_seconds = self.time_to_seconds(time)
        return int((current_seconds / self.total_seconds) * 1000)
    
    def slider_value_to_time(self, value):
        """Конвертация значения слайдера во время"""
        seconds = (value / 1000.0) * self.total_seconds
        return self.seconds_to_time(seconds)
    
    def get_agents_coords_at_time_async(self):
        """Асинхронное получение координат выбранных агентов"""
        # Проверяем кэш
        cache_key = (self.current_time, tuple(sorted(self.selected_agents)))
        if cache_key in self.coord_cache and abs(self.current_time - self.cache_time) < 30:
            return self.coord_cache[cache_key]
        
        # Если есть работающий поток, останавливаем его
        if self.worker_thread and self.worker_thread.isRunning():
            self.worker_thread.terminate()
            self.worker_thread.wait()
        
        # Запускаем новый поток для вычислений
        self.worker_thread = WorkerThread(
            self.agents, 
            self.location_coords, 
            self.current_time, 
            self.base_time, 
            self.selected_agents
        )
        self.worker_thread.result_ready.connect(self.on_coords_calculated)
        self.worker_thread.start()
        
        return []
    
    def on_coords_calculated(self, coords_dict):
        """Обработчик завершения вычисления координат"""
        coords = []
        for agent_id, coord_data in coords_dict.items():
            if coord_data:
                lat, lon = coord_data
                coords.append((agent_id, lat, lon))
        
        # Сохраняем в кэш
        cache_key = (self.current_time, tuple(sorted(self.selected_agents)))
        self.coord_cache[cache_key] = coords
        self.cache_time = self.current_time
        
        # Обновляем метки
        self.coord_label.setText(f"Агентов отображается: {len(coords)}")
        
        # Обновляем маркеры на карте
        if coords or self.show_trajectories:
            time_str = datetime.fromtimestamp(self.current_time).strftime('%d.%m.%Y %H:%M:%S')
            self.update_marker_signal.emit(coords, time_str)
    
    def load_osm_map(self):
        """Загрузка OpenStreetMap с оптимизациями"""
        # Центр карты - Кремль
        initial_lat, initial_lon = 55.751435, 37.617918
        
        html_content = f"""
<!DOCTYPE html>
<html>
<head>
    <title>Траектория движения объектов</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <style>
        html, body, #map {{
            width: 100%;
            height: 100%;
            margin: 0;
            padding: 0;
        }}
        .legend {{
            position: absolute;
            top: 10px;
            right: 10px;
            background: white;
            padding: 10px;
            border-radius: 5px;
            box-shadow: 0 0 15px rgba(0,0,0,0.2);
            z-index: 1000;
            font-family: Arial;
            font-size: 12px;
            max-width: 300px;
        }}
        .agent-marker {{
            transition: transform 0.3s ease;
        }}
        .agent-marker:hover {{
            transform: scale(1.3);
        }}
    </style>
</head>
<body>
    <div id="map"></div>
    <div id="legend" class="legend">
        <b>Легенда</b><br>
        <span id="legend-content">Загрузка...</span>
    </div>
    <script>
        // Глобальные переменные
        var map = null;
        var markers = {{}};
        var polylines = {{}};
        var animationInProgress = false;
        
        // Функция инициализации карты
        function initMap() {{
            // Инициализация карта
            map = L.map('map').setView([{initial_lat}, {initial_lon}], 12);
            
            // Добавляем слой OpenStreetMap
            L.tileLayer('https://{{s}}.tile.openstreetmap.org/{{z}}/{{x}}/{{y}}.png', {{
                attribution: '© OpenStreetMap contributors',
                maxZoom: 19
            }}).addTo(map);
            
            console.log("Карта инициализирована");
            updateLegend("Карта загружена. Выберите агентов для отображения.");
        }}
        
        // Функция для обновления траекторий выбранных агентов
        function updateTrajectories(trajectoriesData) {{
            // Очищаем старые траектории
            for (var agentId in polylines) {{
                map.removeLayer(polylines[agentId]);
            }}
            polylines = {{}};
            
            // Обновляем траектории только если есть данные
            if (trajectoriesData && Object.keys(trajectoriesData).length > 0) {{
                var agentCount = 0;
                for (var agentId in trajectoriesData) {{
                    var points = trajectoriesData[agentId];
                    var color = points.color || '#FF0000';
                    
                    if (points.data && points.data.length > 1) {{
                        // Создаем новую линию - ПРЯМУЮ между точками
                        polylines[agentId] = L.polyline(points.data, {{
                            color: color,
                            weight: 2,
                            opacity: 0.6,
                            smoothFactor: 1.0  // Отключаем сглаживание для прямых линий
                        }}).addTo(map);
                        
                        // Добавляем всплывающую подсказку
                        polylines[agentId].bindPopup("Агент " + agentId);
                        agentCount++;
                    }}
                }}
                updateLegend("Отображается " + agentCount + " траекторий");
            }} else {{
                updateLegend("Траектории не отображаются");
            }}
        }}
        
        // Функция для обновления позиций маркеров с оптимизацией
        window.updateMarkerPositions = function(agentsData, timeStr) {{
            if (animationInProgress) {{
                return; // Пропускаем обновление, если анимация еще в процессе
            }}
            
            // Создаем множество ID агентов, которые должны быть отображены
            var displayIds = {{}};
            for (var i = 0; i < agentsData.length; i++) {{
                displayIds[agentsData[i][0]] = true;
            }}
            
            // Удаляем маркеры для агентов, которых больше нет
            for (var agentId in markers) {{
                if (!displayIds[agentId]) {{
                    map.removeLayer(markers[agentId]);
                    delete markers[agentId];
                }}
            }}
            
            // Обновляем или создаем маркеры с оптимизированной анимацией
            for (var i = 0; i < agentsData.length; i++) {{
                var agentData = agentsData[i];
                var agentId = agentData[0];
                var lat = agentData[1];
                var lon = agentData[2];
                
                if (markers[agentId]) {{
                    // Получаем текущую позицию маркера
                    var currentPos = markers[agentId].getLatLng();
                    
                    // Если позиция изменилась значительно, обновляем
                    if (Math.abs(currentPos.lat - lat) > 0.0001 || Math.abs(currentPos.lng - lon) > 0.0001) {{
                        animationInProgress = true;
                        
                        // Плавно перемещаем маркер к новой позиции с оптимизацией
                        markers[agentId].setLatLng([lat, lon], {{
                            animate: true,
                            duration: 0.3,  // Уменьшили длительность анимации
                            easeLinearity: 0.25
                        }});
                        
                        // Сбрасываем флаг после завершения анимации
                        setTimeout(function() {{
                            animationInProgress = false;
                        }}, 300);
                        
                        // Обновляем всплывающее окно
                        markers[agentId].setPopupContent("<b>Агент " + agentId + "</b><br>Время: " + timeStr + 
                                                       "<br>Координаты: " + lat.toFixed(6) + ", " + lon.toFixed(6));
                    }}
                }} else {{
                    // Цвет маркера в зависимости от ID агента
                    var colors = ['#FF0000', '#00FF00', '#0000FF', '#FFFF00', '#FF00FF', 
                                 '#00FFFF', '#FFA500', '#800080', '#FF69B4', '#32CD32'];
                    var color = colors[agentId % colors.length];
                    
                    // Создаем иконку маркера
                    var markerIcon = L.divIcon({{
                        className: 'agent-marker',
                        html: '<div style="background-color:' + color + '; width:12px; height:12px; ' +
                              'border-radius:50%; border:2px solid white; box-shadow:0 0 3px rgba(0,0,0,0.3);"></div>',
                        iconSize: [12, 12],
                        iconAnchor: [6, 6]
                    }});
                    
                    // Создаем маркер
                    markers[agentId] = L.marker([lat, lon], {{
                        icon: markerIcon,
                        title: 'Агент ' + agentId
                    }}).addTo(map);
                    
                    // Добавляем всплывающее окно
                    markers[agentId].bindPopup("<b>Агент " + agentId + "</b><br>Время: " + timeStr + 
                                             "<br>Координаты: " + lat.toFixed(6) + ", " + lon.toFixed(6));
                }}
            }}
            
            updateLegend("Отображается " + agentsData.length + " агентов");
        }};
        
        // Функция для обновления легенды
        function updateLegend(text) {{
            var legendElement = document.getElementById('legend-content');
            if (legendElement) {{
                legendElement.innerHTML = text;
            }}
        }}
        
        // Инициализируем карту после загрузки страницы
        document.addEventListener('DOMContentLoaded', initMap);
    </script>
</body>
</html>
"""
        
        self.web_view.setHtml(html_content)
    
    def get_selected_trajectory_points(self):
        """Возвращает точки траектории для выбранных агентов (только прямые линии между точками)"""
        if not self.show_trajectories:
            return "{}"
        
        trajectories_js = {}
        
        # Ограничиваем количество агентов для отображения траекторий
        selected_agents = list(self.selected_agents)[:3]  # Еще меньше для производительности
        
        for agent_id in selected_agents:
            if agent_id in self.agents:
                points = []
                agent_data = self.agents[agent_id]
                
                # Добавляем только точки смены локаций (прямые линии между ними)
                for hour, location in agent_data['locations']:
                    if location in self.location_coords:
                        lat, lon = self.location_coords[location]
                        points.append([lat, lon])
                
                if len(points) > 1:
                    trajectories_js[str(agent_id)] = {
                        'data': points,
                        'color': self.agent_colors.get(agent_id, '#FF0000')
                    }
        
        # Преобразуем в JSON-строку
        return json.dumps(trajectories_js)
    
    def on_page_loaded(self):
        """Обработчик загрузки страницы"""
        self.is_page_loaded = True
        print("Страница карты загружена")
        self.update_display()
    
    def update_trajectories(self):
        """Обновление траекторий на карте"""
        if not self.is_page_loaded:
            return
        
        trajectories_js = self.get_selected_trajectory_points()
        js_code = f"""
        if (typeof window.updateTrajectories === 'function') {{
            var trajectoriesData = {trajectories_js};
            window.updateTrajectories(trajectoriesData);
        }}
        """
        self.web_view.page().runJavaScript(js_code)
    
    def update_display(self):
        """Обновление всех элементов интерфейса"""
        if not self.agents or not self.is_page_loaded:
            return
        
        # Преобразуем время в читаемый формат
        time_str = datetime.fromtimestamp(self.current_time).strftime('%d.%m.%Y %H:%M:%S')
        date_str = datetime.fromtimestamp(self.current_time).strftime('%d.%m.%Y')
        time_detail_str = datetime.fromtimestamp(self.current_time).strftime('%H:%M:%S')
        
        self.current_time_label.setText(time_detail_str)
        self.current_date_label.setText(date_str)
        
        # Обновляем значение слайдера
        slider_value = self.time_to_slider_value(self.current_time)
        # Блокируем сигнал, чтобы не вызывать рекурсию
        self.time_slider.blockSignals(True)
        self.time_slider.setValue(slider_value)
        self.time_slider.blockSignals(False)
        
        # Обновляем метку времени
        self.time_label.setText(f"Время: {time_str}")
        
        # Асинхронно получаем координаты агентов
        self.get_agents_coords_at_time_async()
        
        # Обновляем траектории если нужно
        if self.show_trajectories:
            self.update_trajectories()
    
    def update_marker_positions(self, agents_data, time_str):
        """Обновление позиций маркеров через JavaScript"""
        if not self.is_page_loaded:
            return
        
        # Преобразуем данные в формат JS
        agents_json = json.dumps(agents_data)
        
        js_code = f"""
        if (typeof window.updateMarkerPositions === 'function') {{
            window.updateMarkerPositions({agents_json}, "{time_str}");
        }}
        """
        self.web_view.page().runJavaScript(js_code)
    
    def on_slider_changed(self, value):
        """Обработчик изменения слайдера"""
        self.current_time = self.slider_value_to_time(value)
        self.current_seconds = self.time_to_seconds(self.current_time)
        self.update_display()
    
    def toggle_play(self):
        """Включение/выключение автоматического воспроизведения"""
        self.is_playing = not self.is_playing
        
        if self.is_playing:
            self.play_button.setText("⏸")
            # Увеличиваем интервал для лучшей производительности
            self.play_timer.start(50)  # 50 мс вместо 30
        else:
            self.play_button.setText("▶")
            self.play_timer.stop()
    
    def next_frame_step(self):
        """Переход на следующий шаг (30 секунд)"""
        if self.current_seconds < self.total_seconds:
            self.current_seconds += self.frame_step
            self.current_time = self.seconds_to_time(self.current_seconds)
            self.update_display()
    
    def prev_frame_step(self):
        """Переход на предыдущий шаг (30 секунд)"""
        if self.current_seconds > 0:
            self.current_seconds -= self.frame_step
            self.current_time = self.seconds_to_time(self.current_seconds)
            self.update_display()
    
    def reset_to_start(self):
        """Сброс к начальному времени"""
        self.current_seconds = 0
        self.current_time = self.seconds_to_time(self.current_seconds)
        self.update_display()
    
    def next_frame(self):
        """Следующий кадр при воспроизведении"""
        if not self.agents:
            return
        
        # Увеличиваем время с учетом скорости воспроизведения
        step_seconds = self.frame_step * self.play_speed
        self.current_seconds += step_seconds
        
        # Если дошли до конца
        if self.current_seconds > self.total_seconds:
            self.current_seconds = 0
        
        self.current_time = self.seconds_to_time(self.current_seconds)
        self.update_display()

def main():
    app = QApplication(sys.argv)
    window = MapWindow()
    window.show()
    sys.exit(app.exec_())

if __name__ == "__main__":
    main()