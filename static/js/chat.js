/**
 * Chat client for lettersandprompts.com
 * Vanilla JS WebSocket chat implementation
 */

class ChatRoom {
  constructor() {
    this.ws = null;
    this.handle = localStorage.getItem('chat_handle') || '';
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.messageRateLimit = 1000; // 1 second
    this.lastMessageTime = 0;
    
    // DOM elements
    this.modal = document.getElementById('handle-modal');
    this.handleInput = document.getElementById('handle-input');
    this.joinButton = document.getElementById('join-button');
    this.chatInput = document.getElementById('chat-input');
    this.sendButton = document.getElementById('send-button');
    this.messagesContainer = document.getElementById('chat-messages');
    this.userList = document.getElementById('user-list');
    this.userCount = document.getElementById('user-count');
    
    this.init();
  }
  
  init() {
    // Show handle picker if no handle saved
    if (!this.handle) {
      this.showHandlePicker();
    } else {
      this.connect();
    }
    
    // Event listeners
    this.joinButton.addEventListener('click', () => this.joinChat());
    this.handleInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.joinChat();
    });
    
    this.sendButton.addEventListener('click', () => this.sendMessage());
    this.chatInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.sendMessage();
    });
    
    // Auto-focus handle input
    if (!this.handle) {
      this.handleInput.focus();
    }
  }
  
  showHandlePicker() {
    this.modal.style.display = 'flex';
    this.handleInput.focus();
  }
  
  hideHandlePicker() {
    this.modal.style.display = 'none';
  }
  
  joinChat() {
    const handle = this.handleInput.value.trim();
    if (!handle || handle.length > 20) {
      alert('Please enter a valid handle (1-20 characters)');
      return;
    }
    
    // Save handle
    this.handle = handle;
    localStorage.setItem('chat_handle', handle);
    
    this.hideHandlePicker();
    this.connect();
  }
  
  connect() {
    // Determine WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?handle=${encodeURIComponent(this.handle)}`;
    
    console.log('Connecting to', wsUrl);
    
    try {
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.chatInput.disabled = false;
        this.sendButton.disabled = false;
        this.chatInput.focus();
      };
      
      this.ws.onclose = () => {
        console.log('WebSocket closed');
        this.chatInput.disabled = true;
        this.sendButton.disabled = true;
        this.reconnect();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
      
      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (err) {
          console.error('Failed to parse message:', err);
        }
      };
      
    } catch (err) {
      console.error('Failed to create WebSocket:', err);
      this.reconnect();
    }
  }
  
  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.renderSystemMessage('Connection lost. Refresh the page to reconnect.');
      return;
    }
    
    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000);
    
    this.renderSystemMessage(`Connection lost. Reconnecting in ${delay/1000}s...`);
    
    setTimeout(() => this.connect(), delay);
  }
  
  handleMessage(message) {
    switch (message.type) {
      case 'message':
        this.renderMessage(message);
        break;
      case 'system':
        this.renderSystemMessage(message.text);
        break;
      case 'userlist':
        this.updateUserList(message.users);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }
  
  renderMessage(message) {
    const messageEl = document.createElement('div');
    messageEl.className = 'chat-message';
    
    const time = new Date(message.ts * 1000).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });
    
    const timeEl = document.createElement('span');
    timeEl.className = 'chat-time';
    timeEl.textContent = `[${time}]`;
    
    const handleEl = document.createElement('span');
    handleEl.className = 'chat-handle';
    handleEl.style.color = message.color;
    handleEl.textContent = message.handle;
    
    const textEl = document.createElement('span');
    textEl.className = 'chat-text';
    textEl.textContent = message.text;
    
    messageEl.appendChild(timeEl);
    messageEl.appendChild(document.createTextNode(' '));
    messageEl.appendChild(handleEl);
    messageEl.appendChild(document.createTextNode(': '));
    messageEl.appendChild(textEl);
    
    this.messagesContainer.appendChild(messageEl);
    this.scrollToBottom();
  }
  
  renderSystemMessage(text) {
    const messageEl = document.createElement('div');
    messageEl.className = 'chat-system';
    messageEl.textContent = text;
    
    this.messagesContainer.appendChild(messageEl);
    this.scrollToBottom();
  }
  
  updateUserList(users) {
    this.userList.innerHTML = '';
    
    users.sort().forEach(user => {
      const li = document.createElement('li');
      li.textContent = user;
      if (user === this.handle) {
        li.className = 'user-list-self';
      }
      this.userList.appendChild(li);
    });
    
    this.userCount.textContent = users.length;
  }
  
  sendMessage() {
    const text = this.chatInput.value.trim();
    if (!text) return;
    
    // Rate limiting
    const now = Date.now();
    if (now - this.lastMessageTime < this.messageRateLimit) {
      alert('Please slow down. You can send 1 message per second.');
      return;
    }
    this.lastMessageTime = now;
    
    // Check length
    if (text.length > 500) {
      alert('Message too long (max 500 characters)');
      return;
    }
    
    // Send via WebSocket
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = {
        type: 'message',
        text: text
      };
      
      this.ws.send(JSON.stringify(message));
      this.chatInput.value = '';
      this.chatInput.focus();
    } else {
      alert('Not connected. Please wait...');
    }
  }
  
  scrollToBottom() {
    this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
  }
}

// Initialize chat when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => new ChatRoom());
} else {
  new ChatRoom();
}
