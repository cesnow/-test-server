'use client';

import React, {Context, createContext, useContext} from "react";
import {WebSocketCore, WebSocketProtoProps} from "@/components/WebSocket/coreNetwork";

const WebSocketProtoContext: Context<WebSocketProtoProps> = createContext<WebSocketProtoProps>({} as WebSocketProtoProps);

export const WebSocketProvider = ({children}: { children: React.ReactNode }) => {

  const protoCore: WebSocketProtoProps = WebSocketCore();

  return (
    <WebSocketProtoContext.Provider value={protoCore}>
      {children}
    </WebSocketProtoContext.Provider>
  )
}

export const useWebSocket = () => useContext(WebSocketProtoContext);
