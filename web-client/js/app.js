class ChatApp {
    constructor() {
        this.token = localStorage.getItem('token');
        this.user = JSON.parse(localStorage.getItem('user') || '{}');
        this.apiBaseUrl = '/api'; // Use relative path, nginx will proxy to backend
        this.selectedUser = null;
        this.currentMediaFile = null;
        this.websocket = null;
        this.wsReconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        
        // Add page unload event to clean up WebSocket
        window.addEventListener('beforeunload', () => {
            this.disconnectWebSocket();
        });
        
        this.initializeEventListeners();
        this.checkAuth();
    }

    initializeEventListeners() {
        // Auth form
        document.getElementById('authForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleAuth();
        });

        document.getElementById('toggleAuth').addEventListener('click', () => {
            this.toggleAuthMode();
        });

        // Main app functionality
        document.getElementById('logoutBtn').addEventListener('click', () => {
            this.logout();
        });

        document.getElementById('refreshUsersBtn').addEventListener('click', () => {
            this.loadUsers();
        });

        document.getElementById('sendBtn').addEventListener('click', () => {
            this.sendMessage();
        });

        document.getElementById('messageInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // File upload
        document.getElementById('attachBtn').addEventListener('click', () => {
            document.getElementById('fileInput').click();
        });

        document.getElementById('fileInput').addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.handleFileSelect(e.target.files[0]);
            }
        });

        // Message type change
        document.querySelectorAll('input[name="messageType"]').forEach(radio => {
            radio.addEventListener('change', () => {
                this.updateMessageType();
            });
        });

        // History modal
        document.getElementById('viewHistoryBtn').addEventListener('click', () => {
            this.showHistoryModal();
        });

        document.getElementById('closeHistoryBtn').addEventListener('click', () => {
            this.hideHistoryModal();
        });

        document.getElementById('loadHistoryBtn').addEventListener('click', () => {
            this.loadMessageHistory();
        });

        // Message filter
        document.getElementById('messageFilter').addEventListener('change', () => {
            this.loadConversation();
        });

        // Password requirements
        document.getElementById('password').addEventListener('focus', () => {
            this.showPasswordRequirements();
        });

        document.getElementById('password').addEventListener('input', () => {
            this.validatePasswordRealTime();
        });

        document.getElementById('password').addEventListener('blur', () => {
            this.hidePasswordRequirements();
        });

        // Username requirements
        document.getElementById('username').addEventListener('focus', () => {
            this.showUsernameRequirements();
        });

        document.getElementById('username').addEventListener('input', () => {
            this.validateUsernameRealTime();
        });

        document.getElementById('username').addEventListener('blur', () => {
            this.hideUsernameRequirements();
        });
    }

    checkAuth() {
        // Always clear any existing data first
        this.clearAllData();
        
        if (this.token && this.user.id) {
            // Verify token is still valid by making an API call
            this.verifyToken().then(valid => {
                if (valid) {
                    this.showChatInterface();
                    this.loadUsers();
                } else {
                    this.handleInvalidAuth();
                }
            }).catch(() => {
                this.handleInvalidAuth();
            });
        } else {
            this.showAuthModal();
        }
    }

    async handleAuth() {
        const form = document.getElementById('authForm');
        const formData = new FormData(form);
        const isLogin = document.getElementById('authTitle').textContent === 'Login';
        
        // Clear previous errors
        this.clearFormErrors();
        
        const endpoint = isLogin ? '/auth/login' : '/auth/register';
        const data = {
            username: formData.get('username'),
            password: formData.get('password')
        };

        if (!isLogin) {
            data.email = formData.get('email');
        }

        try {
            const response = await this.apiCall(endpoint, {
                method: 'POST',
                body: JSON.stringify(data)
            });

            if (response.success) {
                this.token = response.data.token;
                this.user = response.data.user;
                localStorage.setItem('token', this.token);
                localStorage.setItem('user', JSON.stringify(this.user));
                
                this.hideAuthModal();
                this.showChatInterface();
                this.loadUsers();
                this.showSuccess(response.message);
            } else {
                this.handleFormErrors(response.error || 'Authentication failed');
            }
        } catch (error) {
            console.error('Auth error:', error);
            if (error.message == 'Failed to fetch' || error.name == 'TypeError'){
                this.showFormError('Unable to connect to the server')
            } else if (error.message.includes('Authentication failed')){
                this.showFormError('Session Expired, Try again')

            } else{
                this.showFormError('Unexpected Error Occurred, Try again')
            }
        }
    }

    toggleAuthMode() {
        const title = document.getElementById('authTitle');
        const button = document.getElementById('authButton');
        const toggle = document.getElementById('toggleAuth');
        const emailField = document.getElementById('emailField');
        const passwordRequirements = document.getElementById('passwordRequirements');
        const usernameRequirements = document.getElementById('usernameRequirements');

        if (title.textContent === 'Login') {
            title.textContent = 'Register';
            button.textContent = 'Register';
            toggle.textContent = 'Already have an account? Login';
            emailField.style.display = 'block';
            document.getElementById('email').required = true;
            
            // Show requirements for registration
            if (passwordRequirements) {
                passwordRequirements.style.display = 'block';
            }
            if (usernameRequirements) {
                usernameRequirements.style.display = 'block';
            }
        } else {
            title.textContent = 'Login';
            button.textContent = 'Login';
            toggle.textContent = "Don't have an account? Register";
            emailField.style.display = 'none';
            document.getElementById('email').required = false;
            
            // Hide requirements for login
            if (passwordRequirements) {
                passwordRequirements.style.display = 'none';
            }
            if (usernameRequirements) {
                usernameRequirements.style.display = 'none';
            }
        }
        
        // Clear any existing errors when switching modes
        this.clearFormErrors();
    }

    showAuthModal() {
        // Hide chat interface completely
        this.hideAuthModal(); // First hide to ensure clean state
        document.getElementById('authModal').classList.add('show');
        document.getElementById('chatContainer').style.display = 'none';
        
        // Clear any residual data
        this.clearAllData();
    }

    hideAuthModal() {
        document.getElementById('authModal').classList.remove('show');
    }

    showChatInterface() {
        this.hideAuthModal();
        document.getElementById('chatContainer').style.display = 'flex';
        document.getElementById('userDisplay').textContent = this.user.username;
        document.getElementById('userEmail').textContent = this.user.email;
        document.getElementById('messageInput').disabled = false;
        document.getElementById('sendBtn').disabled = false;
        
        // Connect to WebSocket for real-time messages
        this.connectWebSocket();
    }

    logout() {
        this.clearAllData();
        this.disconnectWebSocket();
        
        this.showInfo('You have been logged out successfully');
        this.showAuthModal();
    }

    // Clear all user data and interface
    clearAllData() {
        this.token = null;
        this.user = {};
        this.selectedUser = null;
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        
        // Clear interface
        this.clearChat();
        this.clearUserDisplay();
    }

    // Handle invalid authentication
    handleInvalidAuth() {
        this.clearAllData();
        this.disconnectWebSocket();
        this.showError('Session expired. Please log in again.');
        this.showAuthModal();
    }

    // Verify token is still valid
    async verifyToken() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });
            return response.ok;
        } catch (error) {
            console.error('Token verification failed:', error);
            return false;
        }
    }

    // Clear user display elements
    clearUserDisplay() {
        document.getElementById('userDisplay').textContent = '';
        document.getElementById('userEmail').textContent = '';
        document.getElementById('usersList').innerHTML = '';
    }

    // WebSocket connection methods
    connectWebSocket() {
        if (!this.token) return;

        try {
            // Connect to WebSocket server through nginx proxy
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(this.token)}`;
            this.websocket = new WebSocket(wsUrl);

            this.websocket.onopen = () => {
                console.log('WebSocket connected');
                this.wsReconnectAttempts = 0;
                this.showInfo('Connected to real-time messaging');
            };

            this.websocket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            this.websocket.onclose = () => {
                console.log('WebSocket disconnected');
                this.websocket = null;
                
                // Attempt to reconnect if not intentionally disconnected
                if (this.token && this.wsReconnectAttempts < this.maxReconnectAttempts) {
                    this.wsReconnectAttempts++;
                    console.log(`Attempting to reconnect WebSocket (${this.wsReconnectAttempts}/${this.maxReconnectAttempts})`);
                    setTimeout(() => this.connectWebSocket(), 3000 * this.wsReconnectAttempts);
                }
            };

            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
        }
    }

    disconnectWebSocket() {
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }
        this.wsReconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    }

    handleWebSocketMessage(data) {
        switch (data.type) {
            case 'new_message':
                this.handleNewMessageNotification(data.data);
                break;
            case 'pong':
                // Handle pong for keepalive
                break;
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    }

    handleNewMessageNotification(message) {
        // Show notification for new messages
        if (message.sender_id !== this.user.id) {
            // It's a message from someone else
            if (message.message_type === 'broadcast') {
                this.showInfo(`New broadcast from ${message.sender_username}: ${message.content.substring(0, 50)}...`);
            } else {
                this.showInfo(`New message from ${message.sender_username}: ${message.content.substring(0, 50)}...`);
            }
        }

        // Update UI if we're viewing the relevant conversation
        if (this.selectedUser && message.message_type === 'direct') {
            // Check if this message is part of current conversation
            if ((message.sender_id === this.selectedUser.id && message.recipient_id === this.user.id) ||
                (message.sender_id === this.user.id && message.recipient_id === this.selectedUser.id)) {
                this.loadConversation(); // Refresh the conversation
            }
        } else if (message.message_type === 'broadcast') {
            // For broadcast messages, just show notification (they'll see it in history)
            // Could optionally refresh if viewing broadcast history
        }
    }

    async loadUsers() {
        try {
            const response = await this.apiCall('/users');
            
            if (response.success) {
                this.displayUsers(response.data.users);
                this.showInfo(`Found ${response.data.users.length} users online`);
            } else {
                this.showError('Failed to load users');
            }
        } catch (error) {
            console.error('Error loading users:', error);
            this.showError('Unable to connect to server. Please check your connection.');
        }
    }

    displayUsers(users) {
        const usersList = document.getElementById('usersList');
        usersList.innerHTML = '';

        users.forEach(user => {
            const li = document.createElement('li');
            li.dataset.userId = user.id;
            li.innerHTML = `
                <div>
                    <div class="user-item-name">${this.escapeHtml(user.username)}</div>
                    <div class="user-item-email">${this.escapeHtml(user.email)}</div>
                </div>
            `;
            li.addEventListener('click', () => this.selectUser(user));
            usersList.appendChild(li);
        });
    }

    selectUser(user) {
        // Update UI
        document.querySelectorAll('#usersList li').forEach(li => li.classList.remove('active'));
        document.querySelector(`#usersList li[data-user-id="${user.id}"]`).classList.add('active');
        
        // Set to direct message mode
        document.querySelector('input[name="messageType"][value="direct"]').checked = true;
        
        this.selectedUser = user;
        this.updateChatTitle();
        this.loadConversation();
    }

    updateMessageType() {
        const messageType = document.querySelector('input[name="messageType"]:checked').value;
        
        if (messageType === 'broadcast') {
            // Clear user selection for broadcast
            document.querySelectorAll('#usersList li').forEach(li => li.classList.remove('active'));
            this.selectedUser = null;
        }
        
        this.updateChatTitle();
        this.clearMessages();
    }

    updateChatTitle() {
        const messageType = document.querySelector('input[name="messageType"]:checked').value;
        const chatTitle = document.getElementById('chatTitle');
        
        if (messageType === 'broadcast') {
            chatTitle.textContent = 'Broadcast Message';
        } else if (this.selectedUser) {
            chatTitle.textContent = `Chat with ${this.selectedUser.username}`;
        } else {
            chatTitle.textContent = 'Select a user for direct message';
        }
    }

    async loadConversation() {
        if (!this.selectedUser) return;

        try {
            const response = await this.apiCall(`/conversations/${this.selectedUser.id}`);
            
            if (response.success) {
                this.displayMessages(response.data);
            } else {
                this.displayMessages('Failed to load conversation');
            }
        } catch (error) {
            console.error('Error loading conversation:', error);
            this.showError('Failed to load conversation');
        }
    }

    displayMessages(messages) {
        const messagesList = document.getElementById('messagesList');
        messagesList.innerHTML = '';

        if (messages.length === 0) {
            messagesList.innerHTML = '<div class="loading">No messages yet. Start the conversation!</div>';
            return;
        }

        messages.forEach(message => {
            this.addMessageToDOM(message);
        });

        this.scrollToBottom();
    }

    addMessageToDOM(message) {
        const messagesList = document.getElementById('messagesList');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${message.message_type}`;
        
        if (message.sender_id === this.user.id) {
            messageDiv.classList.add('own');
        }

        const time = new Date(message.created_at).toLocaleString();
        
        let mediaHtml = '';
        if (message.media_url) {
            if (message.media_type === 'image') {
                mediaHtml = `<div class="message-media"><img src="${message.media_url}" alt="Image" /></div>`;
            } else if (message.media_type === 'video') {
                mediaHtml = `<div class="message-media"><video controls><source src="${message.media_url}" /></video></div>`;
            } else if (message.media_type === 'audio') {
                mediaHtml = `<div class="message-media"><audio controls><source src="${message.media_url}" /></audio></div>`;
            } else {
                mediaHtml = `<div class="message-media"><a href="${message.media_url}" target="_blank">📎 Download File</a></div>`;
            }
        }
        
        messageDiv.innerHTML = `
            <div class="message-header">
                <div>
                    <span class="message-author">${this.escapeHtml(message.sender_username)}</span>
                    <span class="message-type ${message.message_type}">${message.message_type}</span>
                </div>
                <span class="message-time">${time}</span>
            </div>
            <div class="message-content">${this.escapeHtml(message.content)}</div>
            ${mediaHtml}
        `;

        messagesList.appendChild(messageDiv);
    }

    async handleFileSelect(file) {
        if (file.size > 10 * 1024 * 1024) { // 10MB limit
            this.showWarning('File size must be less than 10MB');
            return;
        }

        this.currentMediaFile = file;
        this.showMediaPreview(file);
    }

    showMediaPreview(file) {
        const preview = document.getElementById('mediaPreview');
        const reader = new FileReader();

        reader.onload = (e) => {
            let previewHtml = '';
            
            if (file.type.startsWith('image/')) {
                previewHtml = `<img src="${e.target.result}" alt="Preview" />`;
            } else {
                previewHtml = `<div>📎 ${file.name}</div>`;
            }

            preview.innerHTML = `
                ${previewHtml}
                <div class="preview-info">
                    <div><strong>${file.name}</strong></div>
                    <div>Size: ${this.formatFileSize(file.size)}</div>
                    <button class="remove-media" onclick="chatApp.removeMedia()">Remove</button>
                </div>
            `;
            preview.style.display = 'block';
        };

        reader.readAsDataURL(file);
    }

    removeMedia() {
        this.currentMediaFile = null;
        document.getElementById('mediaPreview').style.display = 'none';
        document.getElementById('fileInput').value = '';
    }

    async uploadMedia(file) {
        const formData = new FormData();
        formData.append('file', file);

        try {
            const response = await fetch(`${this.apiBaseUrl}/media/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                },
                body: formData
            });

            const result = await response.json();
            
            if (result.success) {
                return result.data;
            } else {
                throw new Error(result.error || 'Upload failed');
            }
        } catch (error) {
            console.error('Upload error:', error);
            throw error;
        }
    }

    async sendMessage() {
        const messageInput = document.getElementById('messageInput');
        const content = messageInput.value.trim();
        const messageType = document.querySelector('input[name="messageType"]:checked').value;

        if (!content && !this.currentMediaFile) {
            this.showWarning('Please enter a message or select a file');
            return;
        }

        if (messageType === 'direct' && !this.selectedUser) {
            this.showWarning('Please select a user for direct message');
            return;
        }

        try {
            let mediaData = null;
            
            // Upload media if present
            if (this.currentMediaFile) {
                mediaData = await this.uploadMedia(this.currentMediaFile);
            }

            const messageData = {
                content: content || 'Media file',
                message_type: messageType,
                recipients: messageType === 'direct' ? [this.selectedUser.id] : [],
                media_url: mediaData ? mediaData.url : null,
                media_type: mediaData ? mediaData.media_type : null
            };

            const response = await this.apiCall('/messages', {
                method: 'POST',
                body: JSON.stringify(messageData)
            });

            if (response.success) {
                messageInput.value = '';
                this.removeMedia();
                
                // Add message to DOM if it's relevant to current view
                if (messageType === 'direct' && this.selectedUser) {
                    this.loadConversation(); // Reload conversation
                    this.showInfo('Message sent successfully!');
                } else if (messageType === 'broadcast') {
                    this.showSuccess('Broadcast message sent to all users!');
                }
            } else {
                this.showError(response.error || 'Failed to send message');
            }
        } catch (error) {
            console.error('Error sending message:', error);
            this.showError('Failed to send message');
        }
    }

    showHistoryModal() {
        document.getElementById('historyModal').classList.add('show');
    }

    hideHistoryModal() {
        document.getElementById('historyModal').classList.remove('show');
    }

    async loadMessageHistory() {
        const filter = document.getElementById('historyFilter').value;
        
        try {
            let url = '/messages?limit=100';
            if (filter) {
                url += `&type=${filter}`;
            }

            const response = await this.apiCall(url);
            
            if (response.success) {
                this.displayMessageHistory(response.data.messages);
            } else {
                this.showError('Failed to load message history');
            }
        } catch (error) {
            console.error('Error loading history:', error);
            this.showError('Failed to load message history');
        }
    }

    displayMessageHistory(messages) {
        const historyList = document.getElementById('historyList');
        historyList.innerHTML = '';

        if (messages.length === 0) {
            historyList.innerHTML = '<div class="loading">No messages found</div>';
            return;
        }

        messages.forEach(message => {
            const messageDiv = document.createElement('div');
            messageDiv.className = 'history-item';
            
            const time = new Date(message.created_at).toLocaleString();
            const isOwn = message.sender_id === this.user.id;
            
            messageDiv.innerHTML = `
                <div class="message-header">
                    <div>
                        <span class="message-author">${this.escapeHtml(message.sender_username)}</span>
                        <span class="message-type ${message.message_type}">${message.message_type}</span>
                        ${isOwn ? '<small>(You)</small>' : ''}
                    </div>
                    <span class="message-time">${time}</span>
                </div>
                <div class="message-content">${this.escapeHtml(message.content)}</div>
                ${message.media_url ? `<div><a href="${message.media_url}" target="_blank">📎 Media</a></div>` : ''}
            `;
            
            historyList.appendChild(messageDiv);
        });
    }

    async apiCall(endpoint, options = {}) {
        const url = `${this.apiBaseUrl}${endpoint}`;
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...(this.token && { 'Authorization': `Bearer ${this.token}` })
            }
        };

        try {
            const response = await fetch(url, { ...defaultOptions, ...options });
            
            // Handle authentication errors
            if (response.status === 401 && !endpoint.includes('/auth/')) {
                this.handleInvalidAuth();
                throw new Error('Authentication failed');
            }
            
            return await response.json();
        } catch (error) {
            console.error('API call failed:', error);
            throw error;
        }
    }

    clearChat() {
        document.getElementById('messagesList').innerHTML = '';
        document.getElementById('chatTitle').textContent = 'Select a user or broadcast';
        document.getElementById('messageInput').value = '';
        document.getElementById('messageInput').disabled = true;
        document.getElementById('sendBtn').disabled = true;
        document.getElementById('usersList').innerHTML = '';
        this.selectedUser = null;
        this.currentMediaFile = null;
        
        // Clear any file input
        const fileInput = document.getElementById('fileInput');
        if (fileInput) fileInput.value = '';
    }

    clearMessages() {
        document.getElementById('messagesList').innerHTML = '';
    }

    scrollToBottom() {
        const container = document.getElementById('messagesContainer');
        container.scrollTop = container.scrollHeight;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showWarning(message) {
        this.showNotification(message, 'warning');
    }

    showInfo(message) {
        this.showNotification(message, 'info');
    }

    showNotification(message, type = 'info', duration = 5000) {
        const container = document.getElementById('notificationContainer');
        
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        notification.innerHTML = `
            <div class="notification-content">
                <div class="notification-message">${this.escapeHtml(message)}</div>
                <button class="notification-close" onclick="this.parentElement.parentElement.remove()">&times;</button>
            </div>
            <div class="notification-progress"></div>
        `;
        
        // Add to container
        container.appendChild(notification);
        
        // Trigger animation
        setTimeout(() => {
            notification.classList.add('show');
        }, 10);
        
        // Auto remove after duration
        setTimeout(() => {
            this.removeNotification(notification);
        }, duration);
        
        // Limit number of notifications
        this.limitNotifications(container);
    }

    removeNotification(notification) {
        if (notification && notification.parentElement) {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.remove();
                }
            }, 300);
        }
    }

    limitNotifications(container, maxNotifications = 5) {
        const notifications = container.querySelectorAll('.notification');
        if (notifications.length > maxNotifications) {
            // Remove oldest notifications
            for (let i = 0; i < notifications.length - maxNotifications; i++) {
                this.removeNotification(notifications[i]);
            }
        }
    }

    // New form error handling methods
    clearFormErrors() {
        // Clear general error message
        const errorContainer = document.getElementById('errorContainer');
        if (errorContainer) {
            errorContainer.style.display = 'none';
            errorContainer.textContent = '';
        }

        // Clear field-specific errors
        const fieldErrors = document.querySelectorAll('.field-error');
        fieldErrors.forEach(error => {
            error.classList.remove('show');
            error.textContent = '';
        });

        // Remove error styling from inputs
        const inputs = document.querySelectorAll('.form-group input');
        inputs.forEach(input => {
            input.classList.remove('error');
        });
    }

    showFormError(message) {
        const errorContainer = document.getElementById('errorContainer');
        if (errorContainer) {
            errorContainer.textContent = message;
            errorContainer.style.display = 'block';
        }
    }

    handleFormErrors(errorMessage) {
        // Check if it's a validation error with field information
        if (errorMessage.includes('Field validation for')) {
            this.parseValidationErrors(errorMessage);
        } else {
            // Show general error
            this.showFormError(errorMessage);
        }
    }

    parseValidationErrors(errorMessage) {
        // Parse validation errors like "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"
        
        if (errorMessage.includes("'Password'") && errorMessage.includes("'min'")) {
            this.showFieldError('password', 'Password must be at least 6 characters long');
        } else if (errorMessage.includes("'Username'") && errorMessage.includes("'min'")) {
            this.showFieldError('username', 'Username must be at least 3 characters long');
        } else if (errorMessage.includes("'Email'") && errorMessage.includes("'email'")) {
            this.showFieldError('email', 'Please enter a valid email address');
        } else if (errorMessage.includes("'Username'") && errorMessage.includes("'required'")) {
            this.showFieldError('username', 'Username is required');
        } else if (errorMessage.includes("'Password'") && errorMessage.includes("'required'")) {
            this.showFieldError('password', 'Password is required');
        } else if (errorMessage.includes("'Email'") && errorMessage.includes("'required'")) {
            this.showFieldError('email', 'Email is required');
        } else {
            // Fallback to general error
            this.showFormError(errorMessage);
        }
    }

    showFieldError(fieldName, message) {
        // Show error for specific field
        const errorElement = document.getElementById(fieldName + 'Error');
        const inputElement = document.getElementById(fieldName);
        
        if (errorElement) {
            errorElement.textContent = message;
            errorElement.classList.add('show');
        }
        
        if (inputElement) {
            inputElement.classList.add('error');
        }
    }

    showPasswordRequirements() {
        const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
        const passwordRequirements = document.getElementById('passwordRequirements');
        
        if (isRegisterMode && passwordRequirements) {
            passwordRequirements.classList.add('show');
        }
    }

    hidePasswordRequirements() {
        const passwordRequirements = document.getElementById('passwordRequirements');
        if (passwordRequirements) {
            // Keep showing if in register mode, only hide if switching to login
            const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
            if (!isRegisterMode) {
                passwordRequirements.classList.remove('show');
            }
        }
    }

    validatePasswordRealTime() {
        const password = document.getElementById('password').value;
        const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
        
        if (!isRegisterMode) return;

        // Clear existing field error during typing
        const errorElement = document.getElementById('passwordError');
        const inputElement = document.getElementById('password');
        
        if (password.length >= 6) {
            // Valid password
            if (errorElement) {
                errorElement.classList.remove('show');
                errorElement.textContent = '';
            }
            if (inputElement) {
                inputElement.classList.remove('error');
            }
        } else if (password.length > 0) {
            // Invalid password (but user is typing)
            if (inputElement) {
                inputElement.classList.add('error');
            }
        }
    }

    showUsernameRequirements() {
        const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
        const usernameRequirements = document.getElementById('usernameRequirements');
        
        if (isRegisterMode && usernameRequirements) {
            usernameRequirements.classList.add('show');
        }
    }

    hideUsernameRequirements() {
        const usernameRequirements = document.getElementById('usernameRequirements');
        if (usernameRequirements) {
            const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
            if (!isRegisterMode) {
                usernameRequirements.classList.remove('show');
            }
        }
    }

    validateUsernameRealTime() {
        const username = document.getElementById('username').value;
        const isRegisterMode = document.getElementById('authTitle').textContent === 'Register';
        
        if (!isRegisterMode) return;

        // Clear existing field error during typing
        const errorElement = document.getElementById('usernameError');
        const inputElement = document.getElementById('username');
        
        if (username.length >= 3) {
            // Valid username
            if (errorElement) {
                errorElement.classList.remove('show');
                errorElement.textContent = '';
            }
            if (inputElement) {
                inputElement.classList.remove('error');
            }
        } else if (username.length > 0) {
            // Invalid username (but user is typing)
            if (inputElement) {
                inputElement.classList.add('error');
            }
        }
    }
}

let chatApp;
document.addEventListener('DOMContentLoaded', () => {
    chatApp = new ChatApp();
});
