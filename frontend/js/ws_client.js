const WS_URL = 'ws://localhost:8080/api/voting/ws';

class WSClient {
  constructor() {
    this.socket = null;
    this.handlers = {};
    this.reconnectDelay = 2000;
  }

  connect() {
    this.socket = new WebSocket(WS_URL);

    this.socket.onopen = () => {
      console.log('WebSocket bağlandı.');
      this.reconnectDelay = 2000;
    };

    this.socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (this.handlers[msg.type]) {
          this.handlers[msg.type](msg);
        }
      } catch (e) {
        console.error('WS mesaj hatası:', e);
      }
    };

    this.socket.onclose = () => {
      console.log(`WS kapandı, ${this.reconnectDelay}ms sonra yeniden bağlanıyor…`);
      setTimeout(() => this.connect(), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000);
    };

    this.socket.onerror = (err) => {
      console.error('WS hatası:', err);
    };
  }

  on(type, handler) {
    this.handlers[type] = handler;
    return this;
  }

  disconnect() {
    if (this.socket) {
      this.socket.onclose = null;
      this.socket.close();
    }
  }
}

const ws = new WSClient();