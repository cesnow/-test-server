"use client";

import {useVLProto} from "@/components/TProto/core-network";
import {AuthIdInfo, Ping} from "@/components/TProto/pb/core_types";
import Long from "long";

export default function Home() {

  const {SendHttpMessage} = useVLProto();

  const onSendPing = async () => {
    const pingMsg = await SendHttpMessage(Ping, {pingId: Long.fromNumber(Math.random() * 1e17).toBigInt()});
  }

  const onSendAuthId = async () => {
    const authIdInfo = await SendHttpMessage(AuthIdInfo, {});
  }

  const onSSE = () => {
    // useSSEDefault((data: any) => console.log("SSE", data));
    const eventSource = new EventSource('http://localhost:8801/sse?nt=tylyu&stream=notify')
    eventSource.onmessage = function(event: MessageEvent) {
      console.log(event)
    };
    eventSource.onerror = function() {
      console.log("err");
    };
  }

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-[32px] row-start-2 items-center sm:items-start">
        <div className="flex gap-4 items-center flex-col sm:flex-row">
          <a
            className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"
            onClick={onSendPing}
          >
            Send Http Ping
          </a>
          <a
            className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"
            onClick={onSendAuthId}
          >
            Send Http GetAuthId
          </a>
          <a
            className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"
            onClick={onSSE}
          >
            Connect SSE
          </a>
        </div>
      </main>
      <footer className="row-start-3 flex gap-[24px] flex-wrap items-center justify-center">

      </footer>
    </div>
  );
}
