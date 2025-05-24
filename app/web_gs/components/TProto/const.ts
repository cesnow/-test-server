
export const serverWebSocketUri: string = "ws://localhost:7905";

export type MsgRawData = {
  msgId: bigint;
  seqNo: number;
  event: string;
  body: Buffer;
}
