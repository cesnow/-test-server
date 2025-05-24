'use client';

import {useEffect, useMemo, useRef, useState} from 'react';
import {IncomeMessage, NetMsgData, NetworkData} from "./core-network-message";
import Long from "long";
import useWebSocket, {ReadyState} from "react-use-websocket";
import {serverWebSocketUri} from "./const";
import {
  AuthIdInfo,
  MsgRawData,
  NewSessionCreated,
  Ping,
  Pong,
  RpcError,
  RpcResult,
  TMsgAck,
  TMsgRawDataContainer
} from "@/components/TProto/pb/core_types";
import {Any} from "@/components/TProto/pb/any";
import {getTProtoChecksum, initChecksum, newTObject} from "@/components/TProto/checksum";
import {MessageType} from "@protobuf-ts/runtime";
import axios from "axios";

export interface VLProtoProps {
  ConnectionStatus: string;
  SetSocketUrl: (value: (((prevState: string) => string) | string)) => void;
  SendMessage: (MsgType: MessageType<any>, obj: any) => void
  SendHttpMessage: (MsgType: MessageType<any>, data: any) => Promise<void>
  // SendMessageAsync: <T extends Any>(obj: Any) => Promise<T | null>;
}

export const useVLProto = (): VLProtoProps => {

  const onReceivePong = (obj: Pong): void => {
    console.log(`█████ Pong`, obj);
  }
  const onReceiveNewSessionCreated = (obj: NewSessionCreated): void => {
    console.log(`█████ NewSessionCreated`, obj);
    startPing();
  }
  const onAuthIdInfo = (obj: AuthIdInfo): void => {
    console.log(`█████ AuthIdInfo`, obj);
    NetworkData.getInstance().AuthKeyId = Long.fromBigInt(obj.authId);
    NetworkData.getInstance().SessionId = Long.fromBigInt(obj.sessionId);
    sendMessage(NetworkData.getInstance().MakePackets(
      NetworkData.getInstance().MakeMsgData(Ping, {pingId: Long.fromNumber(999).toBigInt()}).msgData));
  }

  const didUnmount = useRef(false);
  const [socketUrl, setSocketUrl] = useState(serverWebSocketUri);
  const [resolveFunctions, setResolveFunctions] = useState<Record<string, Function>>({});
  const [lastProcessedId, setLastProcessedId] = useState<Long>(Long.MIN_VALUE);

  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startPing = () => {
    if (pingIntervalRef.current) return;
    pingIntervalRef.current = setInterval(() => {
      SendMessage(Ping, {pingId: Long.fromNumber(Math.random() * 1e17).toBigInt()});
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

    let netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(Ping, {pingId: Long.fromNumber(10).toBigInt()});
    let peekKeyPackage = NetworkData.getInstance().MakePackets(netMsg.msgData);
    let peek: Buffer = Buffer.alloc(4);
    peek.writeUInt32BE(0x39F57B94, 0);
    let firstPacket: Buffer = Buffer.concat([peek, peekKeyPackage]);

    sendMessage(firstPacket);
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

  /* System Callbacks
   *  × VLPong
   *  × VLRpcError
   *  × RpcError
   *  × RpcResult
   *  × MsgContainer
   *  × VLNewSessionCreated
   *  × GzipPacked
   */

  const ProcessMessage = <T extends Any>(msgId: Long, sessionId: Long, msg: IncomeMessage): Record<string, IncomeMessage> => {

    switch (msg.raw.typeName) {
      case Pong.typeName:
      case NewSessionCreated.typeName:
      case AuthIdInfo.typeName:
        return {[msg.checksum.toString()]: msg}
      case RpcError.typeName:
        break;
      case RpcResult.typeName:
        let rpcResult: RpcResult = msg.data as RpcResult;
        const incomeMsg: IncomeMessage | null = NetworkData.getInstance().ParseAnyToIncomeMessage(rpcResult.result!)
        if (incomeMsg === null) return {}
        return ProcessMessage(Long.fromBigInt(rpcResult.reqMsgId), sessionId, incomeMsg);
      // case GzipPacked.typeName:
      //   let gp: GzipPacked = msg.raw as GzipPacked;
      //   return ProcessMessage(msgId, sessionId, gp.msg!);
      case TMsgRawDataContainer.typeName:
        const msgContainer: TMsgRawDataContainer = msg.data as TMsgRawDataContainer;
        let result: Record<string, IncomeMessage> = {};
        msgContainer.messages.forEach((msg: MsgRawData) => {
          const buf = Buffer.from(msg.body);
          msg.classId = buf.readInt32LE(0);
          const newMsg = newTObject(msg.classId)
          if (newMsg === null) return;
          result[msg.msgId.toString()] = new IncomeMessage(msg.classId, newMsg, newMsg.fromBinary(buf.subarray(4)));
        });
        return result;
    }
    return {[msgId.toString()]: msg};
  }

  // Receive Message from WebSocket Server
  useEffect((): void => {
    if (lastMessage === null) return;

    let msgData: Blob = lastMessage.data as Blob;
    msgData.arrayBuffer().then((msgArrayBuffer: ArrayBuffer): void => {
      // allSize[4] ( authId[8], sessionId[8], msgId[8], seqNo[4], Bytes[4], Payload[Class[4], Data...] )
      const msgBuffer: Buffer = Buffer.from(msgArrayBuffer);
      const size: number = msgBuffer.readUInt32LE(0);
      if (size > msgArrayBuffer.byteLength - 4) {
        console.error("Size failed");
        return
      }
      console.log(msgBuffer.toString('hex'));
      const payloadBuffer: Buffer = msgBuffer.subarray(4, size + 4);

      let authIdLow = payloadBuffer.readUInt32LE(0);
      let authIdHigh = payloadBuffer.readUInt32LE(4);
      const authId = Long.fromBits(authIdLow, authIdHigh, false);
      let sessionIdLow = payloadBuffer.readUInt32LE(8);
      let sessionIdHigh = payloadBuffer.readUInt32LE(12);
      const sessionId = Long.fromBits(sessionIdLow, sessionIdHigh, false);
      const msgIdLow = payloadBuffer.readUInt32LE(16);
      const msgIdHigh = payloadBuffer.readUInt32LE(20);
      const msgId = Long.fromBits(msgIdLow, msgIdHigh, false);

      const seqNo = payloadBuffer.readUInt32LE(24);
      let bodySize = payloadBuffer.readInt32LE(28);
      let body = Buffer.alloc(bodySize);
      payloadBuffer.copy(body, 0, 32, 32 + bodySize);

      // console.log(body.toString('hex'));

      // if (msgId.lessThanOrEqual(lastProcessedId)) {
      //   console.warn(`Warn: MsgId less than lastProcessedId, msgId:${msgId.toString()} lastProcessedId:${lastProcessedId.toString()}`);
      //   return;
      // }
      console.log(`█████ Received msgId:${msgId.toString()} authId:${authId.toString()} sessionId:${sessionId.toString()} seqNo:${seqNo} bodySize:${bodySize}`);

      let ackMsgIds: Long[] = [];

      if (bodySize > 0) {

        let obj: IncomeMessage | null = NetworkData.getInstance().ParseIncomeMessage(body);
        if (obj == null) {
          console.log("[Common] Error parsing incoming message");
          return;
        }

        const analysisMessages: Record<string, IncomeMessage> = ProcessMessage(msgId, authId, obj);
        // try to execute a system callback
        Object.keys(analysisMessages).forEach(msgIdStr => {
          const replyMessage = analysisMessages[msgIdStr];
          if (msgIdStr in resolveFunctions) {
            resolveFunctions[msgIdStr](replyMessage.data);
            if (replyMessage.raw.typeName === NewSessionCreated.typeName) {
              ackMsgIds.push(msgId);
            }
          } else {
            console.log(`No resolve function for msgId: ${msgIdStr}, type: ${replyMessage.raw.typeName}`, replyMessage.data);
            ackMsgIds.push(msgId);
          }
        });
      }

      setLastProcessedId(msgId);
      if (ackMsgIds.length > 0) {
        const uniqAckIds: Set<Long> = new Set(ackMsgIds);
        Promise.resolve().then(() => {
          // console.log(`VLMsgAckList: ${Array.from(uniqAckIds).join(",")}`);
          let ackIds: bigint[] = []
          uniqAckIds.forEach(v => {
            ackIds.push(v.toBigInt())
          })
          SendMessage(TMsgAck, {msgIds: ackIds})
        });
      }

    });
  }, [lastMessage]);

  /* Export Functions
  *   × SendMessage
  *   × SendMessageAsync
  *   × SetSocketUrl
  *   × ConnectionStatus
  * */
  const SendMessage = (MsgType: MessageType<any>, data: any): void => {
    let netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(MsgType, data);
    let packet: Buffer = NetworkData.getInstance().MakePackets(netMsg.msgData)
    sendMessage(packet);
  }

  // const SendMessageAsync = async <T extends Any>(obj: Any): Promise<T | null> => {
  //   const netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(obj);
  //   const packet: Buffer = NetworkData.getInstance().MakePackets(netMsg.msgData)
  //   sendMessage(packet);
  //
  //   const timeoutPromise = new Promise((_, reject) =>
  //     setTimeout(() => reject(new Error("Request timed out")), 10000)
  //   );
  //
  //   let resolver: (value?: any) => void;
  //   const requestPromise: Promise<T> = new Promise(resolve => {
  //     resolver = resolve;
  //   });
  //   setRequests(prevRequests => ({
  //     ...prevRequests,
  //     [netMsg.msgId.toString()]: {resolve: resolver, promise: requestPromise}
  //   }));
  //
  //   try {
  //     return await Promise.race<T | null>([requestPromise, timeoutPromise]);
  //   } catch (e: any) {
  //     if (!requestPromise.()) {
  //       console.error(e.toString());
  //     }
  //     return null;
  //   } finally {
  //     setRequests(prevRequests => {
  //       const {[netMsg.msgId.toString()]: _, ...otherRequests} = prevRequests;
  //       return otherRequests;
  //     });
  //   }
  //
  // }

  const SendHttpMessage = async (MsgType: MessageType<any>, data: any): Promise<void> => {
    try {
      const netMsg: NetMsgData = NetworkData.getInstance().MakeMsgData(MsgType, data);

      const response = await axios.post("http://localhost:8801/api", netMsg.msgData, {
        headers: {
          "Content-Type": "application/octet-stream",
        },
        responseType: "arraybuffer"
      });

      const payloadBuffer = Buffer.from(response.data);


      let authIdLow = payloadBuffer.readUInt32LE(0);
      let authIdHigh = payloadBuffer.readUInt32LE(4);
      const authId = Long.fromBits(authIdLow, authIdHigh, false);
      let sessionIdLow = payloadBuffer.readUInt32LE(8);
      let sessionIdHigh = payloadBuffer.readUInt32LE(12);
      const sessionId = Long.fromBits(sessionIdLow, sessionIdHigh, false);
      const msgIdLow = payloadBuffer.readUInt32LE(16);
      const msgIdHigh = payloadBuffer.readUInt32LE(20);
      const msgId = Long.fromBits(msgIdLow, msgIdHigh, false);

      const seqNo = payloadBuffer.readUInt32LE(24);
      let bodySize = payloadBuffer.readInt32LE(28);
      let body = Buffer.alloc(bodySize);
      payloadBuffer.copy(body, 0, 32, 32 + bodySize);

      let obj: IncomeMessage | null = NetworkData.getInstance().ParseIncomeMessage(body);
      if (obj == null) {
        console.log("[Common] Error parsing incoming message");
        return;
      }

      const analysisMessages: Record<string, IncomeMessage> = ProcessMessage(msgId, authId, obj);

      Object.keys(analysisMessages).forEach(msgIdStr => {
        const replyMessage = analysisMessages[msgIdStr];
        console.log(resolveFunctions)
        if (msgIdStr in resolveFunctions) {
          resolveFunctions[msgIdStr](replyMessage.data);
        }
      });

      console.log(analysisMessages)

    } catch (error) {
      console.error("Failed to send HTTP message:", error);
    }
  };

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
    initChecksum();
    setResolveFunctions({
      [getTProtoChecksum(Pong).toString()]: onReceivePong,
      [getTProtoChecksum(NewSessionCreated).toString()]: onReceiveNewSessionCreated,
      [getTProtoChecksum(AuthIdInfo).toString()]: onAuthIdInfo
    })
    return () => {
      console.log("█████ WebSocket Dispose");
      NetworkData.getInstance().reset();
    };
  }, []);


  return {
    SendMessage: SendMessage,
    // SendMessageAsync: SendMessageAsync,
    SendHttpMessage: SendHttpMessage,
    SetSocketUrl: SetSocketUrl,
    ConnectionStatus: ConnectionStatus,
  }
};
