import React, { useState, useRef, useEffect } from "react";
import "./App.css";

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [username, setUsername] = useState("");
  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);

  useEffect(() => {
    let storedUsername = localStorage.getItem("username");
    if (!storedUsername) {
      storedUsername =
        prompt("Enter your username:") ||
        "user_" + Math.floor(Math.random() * 1000);
      localStorage.setItem("username", storedUsername);
    }
    setUsername(storedUsername);

    const wsUrl =
      window.location.hostname === "localhost"
        ? "ws://localhost:8000/ws"
        : process.env.REACT_APP_API_URL.replace("http", "ws") + "/ws";
    console.log("Connecting to WebSocket:", wsUrl);
    wsRef.current = new WebSocket(wsUrl);

    wsRef.current.onopen = () => {
      console.log("WebSocket connected");
    };

    wsRef.current.onmessage = (event) => {
      try {
        const { user, text } = JSON.parse(event.data);
        const message = {
          user,
          id: `${Date.now()}-${Math.floor(Math.random() * 1000)}`,
          text,
        };
        setMessages((prevMessages) => [...prevMessages, message].slice(-30));
      } catch (error) {
        console.error("Failed to parse message:", error);
      }
    };

    wsRef.current.onclose = () => {
      console.log("WebSocket disconnected, reconnecting in 5s");
      setTimeout(() => {
        wsRef.current = new WebSocket(wsUrl);
      }, 5000);
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
    if (input.trim() === "") {
      console.log("Input is empty");
      return;
    }
    if (wsRef.current.readyState !== WebSocket.OPEN) {
      console.error(
        "WebSocket not open, readyState:",
        wsRef.current.readyState
      );
      return;
    }
    const message = { user: username, text: input.trim() };
    wsRef.current.send(JSON.stringify(message));
    setInput("");
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
            <strong>{msg.user}:</strong> {msg.text}
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
