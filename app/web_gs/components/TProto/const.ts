
export const serverWebSocketUri: string = "ws://localhost:7905";

export type MsgRawData = {
  msgId: bigint;
  reqMsgId?: bigint;
  event: string;
  body: Buffer;
}
