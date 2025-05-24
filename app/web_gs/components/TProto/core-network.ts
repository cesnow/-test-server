'use client';

import {useEffect, useRef, useState} from 'react';
import {NetMsgData, NetworkData} from "./core-network-message";
import Long from "long";
import useWebSocket, {ReadyState} from "react-use-websocket";
import {MsgRawData, serverWebSocketUri} from "./const";
import {decode, encode} from "@msgpack/msgpack";

export interface TProtoProps {
  ConnectionStatus: string;
  SetSocketUrl: (value: (((prevState: string) => string) | string)) => void;
  SendMessage: (event: string, obj: any) => void
}

export const TProtoCore = (): TProtoProps => {

  const didUnmount = useRef(false);
  const [socketUrl, setSocketUrl] = useState(serverWebSocketUri);
  const [resolveFunctions, setResolveFunctions] = useState<Record<string, Function>>({});
  const [lastProcessedId, setLastProcessedId] = useState<Long>(Long.MIN_VALUE);

  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startPing = () => {
    if (pingIntervalRef.current) return;
    pingIntervalRef.current = setInterval(() => {
      SendMessage("ping", {pingId: 10});
    }, 10_000);
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
    // startPing()
  }

  const {sendMessage, lastMessage, readyState} = useWebSocket(
    socketUrl,
    {
      shouldReconnect: (closeEvent: CloseEvent) => {
        console.log(`ReconnectCode: ${closeEvent.code}`);
        return !didUnmount.current;
      },
      onOpen: onOpen,
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
  }, [lastMessage]);

  const SendMessage = (event: string, data: any): void => {
    let netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(event, data);
    sendMessage(netMsg.msgData);
  }

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
    SendMessage: SendMessage,
    SetSocketUrl: SetSocketUrl,
    ConnectionStatus: ConnectionStatus,
  }
};
