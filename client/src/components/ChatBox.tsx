import { useEffect, useState } from 'react';

interface ChatMessage {
    type: string;
    sender_role?: string;
    content?: string;
    seller_price?: number;
    buyer_price?: number;
    fair_price?: number;
    suggestion?: string;
    error?: string;
}

interface ChatBoxProps {
    role: string;
}

const ChatBox = ({ role }: ChatBoxProps) => {
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [input, setInput] = useState('');

    useEffect(() => {
        if (!role) return;

        const ws = new WebSocket("ws://localhost:8080/ws/chat");

        ws.onopen = () => {
            console.log("WebSocket connected, sending role:", role);
            ws.send(JSON.stringify({ type: 'role', role }));
        };

        ws.onmessage = (event) => {
            try {
                const msg: ChatMessage = JSON.parse(event.data);
                setMessages((prev) => [...prev, msg]);
            } catch (err) {
                console.error("Invalid JSON message:", event.data);
            }
        };

        setSocket(ws);
        return () => ws.close();
    }, [role]);

    const sendMessage = () => {
        if (socket && input.trim() && /^\d+(\.\d+)?$/.test(input)) {
            socket.send(JSON.stringify({ type: 'message', role, content: input }));
            setInput('');
        } else {
            alert("Vui lòng chỉ nhập số để deal giá.");
        }
    };

    return (
        <div className="flex flex-col border rounded p-4 h-96 w-full bg-white">
            <div className="flex-1 overflow-y-auto mb-2">
                {messages.map((msg, idx) => {
                    if (msg.type === 'chat') {
                        return (
                            <div
                                key={idx}
                                className={`mb-1 p-2 rounded ${msg.sender_role === role ? "bg-blue-100 text-right" : "bg-gray-100"}`}
                            >
                                {msg.content}
                            </div>
                        );
                    } else if (msg.type === 'ai_response') {
                        return (
                            <div key={idx} className="mb-1 p-2 rounded bg-green-100 text-center">
                                <div>Giá hợp lý: {msg.fair_price}</div>
                                <div>Gợi ý: {msg.suggestion}</div>
                            </div>
                        );
                    } else if (msg.error) {
                        return (
                            <div key={idx} className="mb-1 p-2 rounded bg-red-100 text-center">
                                {msg.error}
                            </div>
                        );
                    } else {
                        return null;
                    }
                })}
            </div>
            <div className="flex">
                <input
                    className="flex-1 border px-2 py-1 rounded-l"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                    placeholder="Nhập giá muốn đề xuất (chỉ số)"
                />
                <button
                    className="bg-blue-500 text-white px-4 py-1 rounded-r hover:bg-blue-600"
                    onClick={sendMessage}
                >
                    Gửi
                </button>
            </div>
        </div>
    );
};

export default ChatBox;
