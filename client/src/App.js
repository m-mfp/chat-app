import React, { useState, useRef, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import "./App.css";

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [error, setError] = useState("");
  const [typingUsers, setTypingUsers] = useState([]); // New state for typing users
  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);
  const typingTimeoutRef = useRef(null); // Ref to manage typing timeout
  const { loginWithRedirect, logout, user, isAuthenticated, getAccessTokenSilently } = useAuth0();

  // Store Auth0 token and username in localStorage
  useEffect(() => {
    if (isAuthenticated && user) {
      const storeToken = async () => {
        try {
          const token = await getAccessTokenSilently();
          localStorage.setItem("authToken", token);
          localStorage.setItem("username", user.nickname || user.email || "user");
        } catch (error) {
          setError("Failed to get Auth0 token");
        }
      };
      storeToken();
    }
  }, [isAuthenticated, user, getAccessTokenSilently]);

  // Clear localStorage on page unload
  useEffect(() => {
    const handleUnload = () => {
      localStorage.removeItem("authToken");
      localStorage.removeItem("username");
    };
    window.addEventListener("unload", handleUnload);
    return () => window.removeEventListener("unload", handleUnload);
  }, []);

  // Connect to WebSocket
  useEffect(() => {
    if (isAuthenticated) {
      const wsUrl =
        window.location.hostname === "localhost"
          ? "ws://localhost:8000/ws"
          : process.env.REACT_APP_API_URL.replace("http", "ws") + "/ws";
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log("WebSocket connected");
      };

      wsRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "message") {
            const message = {
              user: data.user,
              id: `${Date.now()}-${Math.floor(Math.random() * 1000)}`,
              text: data.text,
            };
            setMessages((prevMessages) => [...prevMessages, message].slice(-30));
          } else if (data.type === "typing") {
            setTypingUsers((prev) => {
              if (!prev.includes(data.user)) {
                return [...prev, data.user];
              }
              return prev;
            });
            // Clear typing indicator after 2 seconds
            clearTimeout(typingTimeoutRef.current);
            typingTimeoutRef.current = setTimeout(() => {
              setTypingUsers((prev) => prev.filter((u) => u !== data.user));
            }, 2000);
          }
        } catch (error) {
          setError("Failed to parse message");
        }
      };

      wsRef.current.onclose = () => {
        setTimeout(() => {
          if (isAuthenticated) {
            wsRef.current = new WebSocket(wsUrl);
          }
        }, 5000);
      };

      wsRef.current.onerror = () => {
        setError("WebSocket connection error");
      };

      return () => {
        wsRef.current.close();
      };
    }
  }, [isAuthenticated]);

  // Scroll to latest message
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  // Handle typing event
  const handleTyping = () => {
    if (wsRef.current.readyState === WebSocket.OPEN) {
      const username = user.nickname || user.email || "user";
      wsRef.current.send(JSON.stringify({ type: "typing", user: username }));
    }
  };

  // Handle send message
  const handleSend = async (e) => {
    e.preventDefault();
    if (input.trim() === "") {
      setError("Message cannot be empty");
      return;
    }
    if (wsRef.current.readyState !== WebSocket.OPEN) {
      setError("WebSocket not connected");
      return;
    }
    const username = user.nickname || user.email || "user";
    const message = { type: "message", user: username, text: input.trim() };
    wsRef.current.send(JSON.stringify(message));
    setInput("");
  };

  // Render auth or chat UI
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white p-8 rounded-lg shadow-lg w-full max-w-md">
          <h2 className="text-2xl font-bold mb-6 text-center">Omni Chat App</h2>
          {error && <p className="text-red-500 mb-4">{error}</p>}
          <button
            onClick={() => loginWithRedirect()}
            className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600"
          >
            Login with Auth0
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="App min-h-screen bg-gray-100 flex flex-col items-center p-4">
      <div className="w-full max-w-2xl">
        <div className="flex justify-between items-center mb-4">
          <h1 className="text-2xl font-bold">
            <img
              id="h1-logo"
              src={process.env.PUBLIC_URL + "/logo192.jpg"}
              alt="Logo"
              className="inline-block w-8 h-8 mr-2"
            />
            Omni Chat App
          </h1>
          <button
            onClick={() => logout({ returnTo: window.location.origin })}
            className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
          >
            Logout
          </button>
        </div>
        {error && <p className="text-red-500 mb-4">{error}</p>}
        {typingUsers.length > 0 && (
          <p className="text-gray-500 mb-2">
            {typingUsers.join(", ")} {typingUsers.length > 1 ? "are" : "is"} typing...
          </p>
        )}
        <div className="chat-box bg-white p-4 rounded-lg shadow-lg h-96 overflow-y-auto mb-4">
          {messages.map((msg) => (
            <div key={msg.id} className="chat-message mb-2">
              <strong>{msg.user}:</strong> {msg.text}
            </div>
          ))}
          <div ref={messagesEndRef}></div>
        </div>
        <form onSubmit={handleSend} className="chat-input flex gap-2">
          <input
            type="text"
            placeholder="Type your message..."
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              handleTyping();
            }}
            className="flex-1 p-2 border rounded"
          />
          <button
            type="submit"
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}

export default App;