import React, { useState, useRef, useEffect } from "react";
import "./App.css";

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);

  useEffect(() => {
    // Connect to WebSocket
    const wsUrl = process.env.REACT_APP_API_URL.replace("http", "ws") + "/ws";
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      console.log("WebSocket connected");
    };

    wsRef.current.onmessage = (event) => {
      const message = {
        user: "user", // Replace with actual user ID in production
        id: `${Date.now()}-${Math.floor(Math.random() * 1000)}`,
        text: event.data,
      };
      setMessages((prevMessages) => [...prevMessages, message].slice(-30));
    };

    wsRef.current.onclose = () => {
      console.log("WebSocket disconnected");
    };

    wsRef.current.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    return () => {
      wsRef.current.close();
    };
  }, []);

  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (input.trim() !== "" && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(input.trim());
      setInput("");
    }
  };

  return (
    <div className="App">
      <h1>
        <img
          id="h1-logo"
          src={process.env.PUBLIC_URL + "/logo192.jpg"}
          alt="Logo"
        ></img>{" "}
        Omni Chat App
      </h1>

      <div className="chat-box">
        {messages.map((msg) => (
          <div key={msg.id} className="chat-message">
            {msg.text}
          </div>
        ))}
        <div ref={messagesEndRef}></div>
      </div>

      <form onSubmit={handleSend} className="chat-input">
        <input
          type="text"
          placeholder="Type your message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <button type="submit">Send</button>
      </form>
    </div>
  );
}

export default App;