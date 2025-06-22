// web/static/app.js

// 等待DOM内容完全加载后再执行脚本
document.addEventListener('DOMContentLoaded', () => {
    // 获取页面上的元素
    const taskInputField = document.getElementById('task-input-field');
    const addTaskBtn = document.getElementById('add-task-btn');
    const taskList = document.getElementById('task-list');
    const timeDisplay = document.querySelector('.time-display');
    const startBtn = document.getElementById('start-btn');
    const pauseBtn = document.getElementById('pause-btn');
    const resetBtn = document.getElementById('reset-btn');

    // --- WebSocket 初始化 ---
    // 建立到服务器的WebSocket连接
    // 'ws://' + window.location.host + '/ws' 会动态地创建正确的WebSocket地址
    // 例如 ws://localhost:8080/ws
    const socket = new WebSocket('ws://' + window.location.host + '/ws');

    socket.onopen = () => {
        console.log('WebSocket连接已建立');
    };

    socket.onclose = () => {
        console.log('WebSocket连接已断开');
    };

    // 监听从服务器发来的消息
    socket.onmessage = (event) => {
        // 后端发送的是Go的Duration字符串，例如 "24m59s"
        const durationString = event.data;
        // 我们需要一个函数来格式化它
        timeDisplay.textContent = formatDuration(durationString);
    };

    socket.onerror = (error) => {
        console.error('WebSocket 错误:', error);
    };

    // --- 函数定义 ---

    // 格式化Go的Duration字符串为 MM:SS
    const formatDuration = (durationStr) => {
        // 使用正则表达式匹配分钟和秒
        const minMatch = durationStr.match(/(\d+)m/);
        const secMatch = durationStr.match(/(\d+)s/);
        
        const minutes = minMatch ? parseInt(minMatch[1], 10) : 0;
        const seconds = secMatch ? parseInt(secMatch[1], 10) : 0;

        // 格式化为两位数
        const paddedMinutes = String(minutes).padStart(2, '0');
        const paddedSeconds = String(seconds).padStart(2, '0');

        return `${paddedMinutes}:${paddedSeconds}`;
    };

    // 从后端API获取任务并渲染到页面上
    const fetchAndRenderTasks = async () => {
        try {
            // 使用fetch API向后端发送GET请求
            const response = await fetch('/api/tasks');
            if (!response.ok) {
                throw new Error('网络请求失败');
            }
            const tasks = await response.json();

            // 清空当前的任务列表
            taskList.innerHTML = '';

            // 如果没有任务，则不进行任何操作
            if (!tasks) return;

            // 遍历从后端获取的任务数组
            tasks.forEach(task => {
                // 为每个任务创建一个列表项
                const listItem = document.createElement('li');
                listItem.textContent = `[ID: ${task.id}] ${task.description}`;
                // 将列表项添加到任务列表中
                taskList.appendChild(listItem);
            });
        } catch (error) {
            console.error('获取任务失败:', error);
        }
    };

    // 添加一个新任务
    const addTask = async () => {
        const description = taskInputField.value.trim();
        if (!description) {
            alert('请输入任务描述！');
            return;
        }

        try {
            // 发送POST请求到后端API
            await fetch('/api/tasks', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ description: description }),
            });

            // 清空输入框
            taskInputField.value = '';
            // 重新获取并渲染任务列表，以显示新添加的任务
            fetchAndRenderTasks();
        } catch (error) {
            console.error('添加任务时出错:', error);
        }
    };

    // --- 事件监听 ---

    // 计时器控制
    startBtn.addEventListener('click', () => socket.send('start'));
    pauseBtn.addEventListener('click', () => socket.send('pause'));
    resetBtn.addEventListener('click', () => socket.send('reset'));

    // 任务管理
    addTaskBtn.addEventListener('click', addTask);
    
    // 也可以通过按Enter键添加任务
    taskInputField.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addTask();
        }
    });

    // --- 初始加载 ---

    // 页面加载时，立即获取并显示任务
    fetchAndRenderTasks();
}); 