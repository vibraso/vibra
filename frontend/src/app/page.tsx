"use client"

import { useState, useEffect } from "react";
import Image from "next/image";
import { useUser } from "./context/UserContext";

interface Author {
  fid: number;
  username: string;
  display_name: string;
  pfp_url: string;
  followerCount: number;
}

interface Reply {
  hash: string;
  author: {
    fid: number;
    username: string;
    display_name: string;
    pfp_url: string;
  };
  text: string;
  timestamp: string;
  reactions: {
    likes_count: number;
    recasts_count: number;
  };
}

interface LivestreamData {
  castHash: string;
  streamer: Author;
  streamUrl: string;
  text: string;
  timestamp: string;
  likes_count: number;
  recasts_count: number;
  replies_count: number;
  channel: {
    id: string;
    name: string;
    image_url: string;
  };
}

interface ApiResponse {
  livestreamData: LivestreamData;
  repliesToLivestream: Reply[];
}

export default function Home() {
  const { user, login, checkSignerStatus } = useUser();
  const [data, setData] = useState<ApiResponse | null>(null);
  const [replyText, setReplyText] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/present`);
        const result = await response.json();
        console.log("the response is", result);
        setData(result);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchData();
  }, []);

  const handleReplySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!replyText.trim() || isSubmitting) return;

    setIsSubmitting(true);

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/cast`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ text: replyText }),
      });

      if (!response.ok) {
        throw new Error('Failed to submit cast');
      }

      const result = await response.json();
      console.log("Cast submitted successfully:", result);

      // Clear the textarea after submitting
      setReplyText("");

      // Optionally, you can update the local state to show the new reply immediately
      if (data) {
        const newReply: Reply = {
          hash: result.hash || Date.now().toString(), // Use the hash from the response if available
          author: {
            fid: result.author?.fid || 0,
            username: result.author?.username || "anon_user",
            display_name: result.author?.display_name || "Anon",
            pfp_url: result.author?.pfp_url || "https://wrpcd.net/cdn-cgi/image/anim=false,fit=contain,f=auto,w=168/https%3A%2F%2Fwarpcast.com%2Favatar.png",
          },
          text: replyText,
          timestamp: result.timestamp || new Date().toISOString(),
          reactions: {
            likes_count: 0,
            recasts_count: 0,
          },
        };
        setData({
          ...data,
          repliesToLivestream: [newReply, ...data.repliesToLivestream],
        });
      }
    } catch (error) {
      console.error("Error submitting cast:", error);
      // Handle error (e.g., show an error message to the user)
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!data) {
    return <div className="flex h-screen items-center justify-center">
      <div className="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-purple-500"></div>
    </div>;
  }

  const { livestreamData, repliesToLivestream } = data;

  // Sort replies by timestamp, most recent first
  const sortedReplies = [...repliesToLivestream].sort((a, b) => 
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );

  // Create a Set of unique participant FIDs
  const uniqueParticipants = new Set(repliesToLivestream.map(reply => reply.author.fid));

  return (
    <div className="flex w-screen flex-col h-screen bg-gray-100">
      {/* Top rectangle with YouTube video */}
      <div className="w-full aspect-video bg-black">
        <iframe 
          className="w-full h-full"      
          src="https://www.youtube.com/embed/dZsIQV-B9Us?si=5rk35HAp24sof9HL" 
          title="YouTube video player" 
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" 
          referrerPolicy="strict-origin-when-cross-origin" 
          allowFullScreen
        ></iframe>
      </div>

      {/* Social media interface layout */}
      <div className="flex-grow flex overflow-hidden">
        {/* Left sidebar */}
        <div className="md:flex flex-col hidden w-1/4 bg-white p-4 overflow-y-auto shadow-lg">
          <h2 className="text-2xl font-bold mb-4 text-purple-600">Livestream Info</h2>
          <div className="mb-4">
            <Image
              src={livestreamData.streamer.pfp_url}
              alt={livestreamData.streamer.display_name}
              width={80}
              height={80}
              className="rounded-full mb-2"
            />
            <h3 className="text-xl font-semibold text-gray-800">{livestreamData.streamer.display_name}</h3>
            <p className="text-gray-600">@{livestreamData.streamer.username}</p>
            <p className="text-sm text-gray-500">{livestreamData.streamer.followerCount} followers</p>
          </div>
          <div className="mb-4">
            <h4 className="font-semibold text-gray-800">Channel</h4>
            <div className="flex items-center">
              <Image
                src={livestreamData.channel.image_url}
                alt={livestreamData.channel.name}
                width={24}
                height={24}
                className="rounded mr-2"
              />
              <span className="text-gray-700">{livestreamData.channel.name}</span>
            </div>
          </div>
          <div className="mb-4">
            <h4 className="font-semibold text-gray-800">Engagement</h4>
            <p className="text-gray-700">👍 {livestreamData.likes_count} likes</p>
            <p className="text-gray-700">🔁 {livestreamData.recasts_count} recasts</p>
            <p className="text-gray-700">💬 {livestreamData.replies_count} replies</p>
          </div>
          <div>
            <h4 className="font-semibold text-gray-800">Cast</h4>
            <p className="text-sm text-gray-700">{livestreamData.text}</p>
            <p className="text-xs text-gray-500 mt-1">
              {new Date(livestreamData.timestamp).toLocaleString()}
            </p>
          </div>
        </div>

        {/* Main content area */}
        <div className="w-full md:w-1/2 bg-gray-50 p-4 overflow-y-auto">
          
          <form onSubmit={handleReplySubmit} className="mb-6 bg-purple-200 p-4 rounded-lg">
      <textarea
        value={replyText}
        onChange={(e) => setReplyText(e.target.value)}
        placeholder="Welcome to the chat..."
        className="w-full p-2 border border-gray-300 rounded-lg text-black focus:outline-none focus:ring-2 focus:ring-purple-600"
        rows={3}
      />
      <div className="flex items-center justify-between mt-2">
        <label htmlFor="image-upload" className="cursor-pointer">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <input
            id="image-upload"
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              if (e.target.files && e.target.files[0]) {
                console.log("File selected:", e.target.files[0]);
              }
            }}
          />
        </label>
        <button
          type="submit"
          className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-600 focus:ring-opacity-50"
          disabled={isSubmitting}
        >
          {isSubmitting ? "Casting..." : "Cast"}
        </button>
      </div>
    </form>

          {sortedReplies.map((reply) => (
            <div key={reply.hash} className="mb-4 p-4 bg-white rounded-lg shadow">
              <div className="flex items-center mb-2">
                <Image
                  src={reply.author.pfp_url}
                  alt={reply.author.display_name}
                  width={40}
                  height={40}
                  className="rounded-full mr-2"
                />
                <div>
                  <p className="font-bold text-purple-700">{reply.author.display_name}</p>
                  <p className="text-sm text-gray-500">@{reply.author.username}</p>
                </div>
              </div>
              <p className="mb-2 text-black">{reply.text}</p>
              <div className="flex justify-between text-sm text-gray-500">
                <span>{new Date(reply.timestamp).toLocaleString()}</span>
                <div>
                  <span className="mr-2">❤️ {reply.reactions.likes_count}</span>
                  <span>🔁 {reply.reactions.recasts_count}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Right sidebar */}
        <div className="hidden md:w-1/4 bg-white p-4 overflow-y-auto shadow-lg">
          <h2 className="text-2xl font-bold mb-4 text-purple-600">Participants</h2>
          <div className="space-y-2">
            {Array.from(uniqueParticipants).map((fid) => {
              const participant = repliesToLivestream.find(reply => reply.author.fid === fid)!.author;
              return (
                <div key={fid} className="flex items-center text-black">
                  <Image
                    src={participant.pfp_url}
                    alt={participant.display_name}
                    width={32}
                    height={32}
                    className="rounded-full mr-2"
                  />
                  <span>{participant.display_name}</span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="w-full h-16 bg-gradient-to-r from-red-500 via-yellow-500 via-green-500 via-blue-500 to-purple-500 flex items-center justify-center">
        <h1 className="text-4xl font-bold text-white animate-pulse">vibra</h1>
      </div>
    </div>
  );
}