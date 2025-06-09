"use client";

import {useWebSocket} from "@/components/WebSocket/provider";
import {useEffect, useState} from "react";
import {events} from "@/lib/sample-data";
import {Header} from "@/components/Layout/header";
import {Sidebar} from "@/components/Layout/sidebar";
import {EventDetails} from "@/components/Layout/event-details";

export default function Home() {

  const [selectedEvent, setSelectedEvent] = useState(events[0]);
  const [searchQuery, setSearchQuery] = useState('');

  const filteredEvents = events.filter(event =>
    event.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    event.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    event.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
    <>
      <div className="min-h-screen bg-background">
        <Header searchQuery={searchQuery} setSearchQuery={setSearchQuery}/>
        <div className="flex">
          <Sidebar
            events={filteredEvents}
            selectedEvent={selectedEvent}
            onEventSelectAction={setSelectedEvent}
          />
          <main className="flex-1 min-h-[calc(100vh-4rem)]">
            <EventDetails event={selectedEvent}/>
          </main>
        </div>
      </div>
    </>
    // {/*<main className="flex flex-col gap-[32px] row-start-2 items-center sm:items-start">*/}
    // {/*  <div className="flex gap-4 items-center flex-col sm:flex-row">*/}
    // {/*    <a*/}
    // {/*      className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"*/}
    // {/*      onClick={() => onSendPing()}*/}
    // {/*    >*/}
    // {/*      Ping*/}
    // {/*    </a>*/}
    // {/*    <a*/}
    // {/*      className="rounded-full border border-solid border-transparent transition-colors flex items-center justify-center bg-foreground text-background gap-2 hover:bg-[#383838] dark:hover:bg-[#ccc] font-medium text-sm sm:text-base h-10 sm:h-12 px-4 sm:px-5 sm:w-auto"*/}
    // {/*      onClick={() => onTest()}*/}
    // {/*    >*/}
    // {/*      Test*/}
    // {/*    </a>*/}
    // {/*  </div>*/}
    // {/*</main>*/}
    // {/*<footer className="row-start-3 flex gap-[24px] flex-wrap items-center justify-center"></footer>
    // */}
  );
}
