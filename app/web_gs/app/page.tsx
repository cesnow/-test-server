"use client";

import {useWebSocket} from "@/components/WebSocket/provider";
import {useEffect} from "react";

export default function Home() {

  const {SendMessage, onEvent} = useWebSocket();

  const onSendPing = () => {
    SendMessage("ping", {pingId: 123});
  }

  const onTest = () => {
    SendMessage("test", {"hi": "199"}, (data: any) => {
      console.log("test data on callback", data);
    });
  }

  useEffect(() => {
    onEvent("test", (data: any) => {
      console.log("test data on event", data);
    })
  }, [])

  return (
    <div
      className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-[32px] row-start-2 items-center sm:items-start">
        <div className="flex gap-4 items-center flex-col sm:flex-row">
          <a
            className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"
            onClick={() => onSendPing()}
          >
            Ping
          </a>
          <a
            className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"
            onClick={() => onTest()}
          >
            Test
          </a>
        </div>
      </main>
      <footer className="row-start-3 flex gap-[24px] flex-wrap items-center justify-center"></footer>
    </div>
  );
}
