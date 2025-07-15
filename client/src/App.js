import React, { useState } from 'react';
import './App.css';

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');

  const handleSend = () => {
    if (input.trim() !== '') {
      const newMessage = {
        id: messages.length + 1,
        text: input.trim(),
      };
      setMessages([...messages, newMessage]);
      setInput(''); // Clear the input box
    }
  };

  return (
    <div className="App">
      <h1>💬 Simple Chat App</h1>

      {/* Message List */}
      <div className="chat-box">
        {messages.map((msg) => (
          <div key={msg.id} className="chat-message">
            {msg.text}
          </div>
        ))}
      </div>

      {/* Input Box */}
      <div className="chat-input">
        <input
          type="text"
          placeholder="Type your message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <button onClick={handleSend}>Send</button>
      </div>
    </div>
  );
}

export default App;
