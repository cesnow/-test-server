'use client';

import {useCallback, useEffect, useRef, useState} from 'react';
import {NetMsgData, NetworkData} from "./coreNetworkMessage";
import Long from "long";
import useWebSocket, {ReadyState} from "react-use-websocket";
import {MsgRawData, serverWebSocketUri} from "./const";
import {decode, encode} from "@msgpack/msgpack";

export interface WebSocketProtoProps {
  ConnectionStatus: string;
  SetSocketUrl: (value: (((prevState: string) => string) | string)) => void;
  SendMessage: (event: string, obj: any, callback?: Function) => void
  onEvent: (eventName: string, callback: Function) => void
}

export const WebSocketCore = (): WebSocketProtoProps => {

  const didUnmount = useRef(false);
  const [socketUrl, setSocketUrl] = useState(serverWebSocketUri);
  const [resolveFunctions, setResolveFunctions] = useState<Record<string, Function>>({});
  const [lastProcessedId, setLastProcessedId] = useState<Long>(Long.MIN_VALUE);
  const [eventListeners, setEventListeners] = useState<Record<string, Function[]>>({});
  const [responseResolvers, setResponseResolvers] = useState<Record<string, Function>>({});

  const onEvent = (eventName: string, callback: Function) => {
    setEventListeners(prev => {
      const callbacks = prev[eventName] || [];
      return {...prev, [eventName]: [...callbacks, callback]};
    });
  };

  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startPing = () => {
    if (pingIntervalRef.current) return;
    pingIntervalRef.current = setInterval(() => {
      SendMessage("ping", {pingId: 10});
    }, 30_000);
  };

  const stopPing = () => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
  };

  const onOpen = (event: Event) => {
    console.log("█████ WebSocket connected!");
    NetworkData.getInstance().reset();
    console.log("█████ Start Peek Session");
    let netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData("ping", {
      ping: Long.fromNumber(10015).toBigInt()}
    );
    console.log(netMsg.msgData.toString('hex'));
    sendMessage(netMsg.msgData);
    startPing()
  }

  const onClose = (event: CloseEvent) => {
    console.log("█████ WebSocket closed!");
    stopPing()
  }

  const {sendMessage, lastMessage, readyState} = useWebSocket(
    socketUrl,
    {
      shouldReconnect: (closeEvent: CloseEvent) => {
        console.log(`ReconnectCode: ${closeEvent.code}`);
        return !didUnmount.current;
      },
      onOpen: onOpen,
      onClose: onClose,
      reconnectAttempts: 30,
      reconnectInterval: 5000,
    }
  );

  // Receive Message from WebSocket Server
  useEffect((): void => {
    if (lastMessage === null) return;

    let msgData: Blob = lastMessage.data as Blob;
    msgData.arrayBuffer().then((msgArrayBuffer: ArrayBuffer): void => {
      const msgBuffer: Buffer = Buffer.from(msgArrayBuffer);

      console.log(msgBuffer.toString('hex'))

      const data = decode<MsgRawData>(msgBuffer, {useBigInt64: true, context: {} as MsgRawData}) as MsgRawData;
      console.log(data);

      const payload = decode(data.body, {useBigInt64: true});
      console.log(payload)

      if (data.reqMsgId && responseResolvers[data.reqMsgId.toString()]) {
        const reqId = data.reqMsgId.toString();
        console.log(`Resolve ${reqId}`);
        responseResolvers[reqId](payload);
        setResponseResolvers(prev => {
          const next = {...prev};
          delete next[reqId];
          return next;
        });
      }

      if (data.event && eventListeners[data.event]) {
        eventListeners[data.event].forEach(cb => cb(payload));
      }

      let ackMsgIds: Long[] = [];

      // setLastProcessedId(msgId);
      // if (ackMsgIds.length > 0) {
      //   const uniqAckIds: Set<Long> = new Set(ackMsgIds);
      //   Promise.resolve().then(() => {
      //     // console.log(`VLMsgAckList: ${Array.from(uniqAckIds).join(",")}`);
      //     let ackIds: bigint[] = []
      //     uniqAckIds.forEach(v => {
      //       ackIds.push(v.toBigInt())
      //     })
      //     SendMessage({msgIds: ackIds})
      //   });
      // }

    });
  }, [lastMessage, eventListeners, responseResolvers]);

  const SendMessage = useCallback(
    (event: string, data: any, callback?: Function): void => {
      const netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(event, data);
      const msgIdStr = netMsg.msgId.toString();

      if (callback) {
        setResponseResolvers(prev => ({
          ...prev,
          [msgIdStr]: callback
        }));
      }

      sendMessage(netMsg.msgData);
    }, [sendMessage, setResponseResolvers]
  )

  const SetSocketUrl = setSocketUrl;
  const ConnectionStatus: string = {
    [ReadyState.CONNECTING]: 'Connecting',
    [ReadyState.OPEN]: 'Open',
    [ReadyState.CLOSING]: 'Closing',
    [ReadyState.CLOSED]: 'Closed',
    [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
  }[readyState];

  // WebSocket Dispose
  useEffect(() => {
    setResolveFunctions({
    })
    return () => {
      console.log("█████ WebSocket Dispose");
      NetworkData.getInstance().reset();
    };
  }, []);


  return {
    SendMessage,
    SetSocketUrl,
    ConnectionStatus,
    onEvent,
  };
};
