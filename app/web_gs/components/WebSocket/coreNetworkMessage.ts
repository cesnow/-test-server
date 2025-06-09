import {encode} from "@msgpack/msgpack";
import Long from "long";

export class NetMsgData {
  msgId: Long;
  msgData: Buffer;

  constructor(msgId: Long, msgData: Buffer) {
    this.msgId = msgId;
    this.msgData = msgData;
  }
}

export class NetworkData {
  private static instance: NetworkData;

  private seq: number = 0;

  private constructor() {
  }

  public reset(): void {
    this.seq = 0;
  }

  public static getInstance(): NetworkData {
    if (!NetworkData.instance) {
      NetworkData.instance = new NetworkData();
    }
    return NetworkData.instance;
  }

  public NextMessageId(req: boolean): Long {
    const unixNano = Long.fromNumber(Date.now()).multiply(1_000_000);
    const ts = unixNano.div(1_000_000_000);
    const ms = unixNano.mod(1_000_000_000).div(1_000_000);
    this.seq = (this.seq + 1) & 0x1ffff;
    if (this.seq === 0x1ffff) this.seq = 0;
    const sid = Long.fromNumber(this.seq);
    const last = Long.fromNumber(req ? 1 : 3);
    return ts.shiftLeft(32)
      .or(ms.shiftLeft(21))
      .or(sid.shiftLeft(3))
      .or(last);
  }

  public MakeMsgData(event: string, data: object): NetMsgData {
    let msgId: Long = this.NextMessageId(true);
    const buffer = {
      msgId: msgId.toBigInt(),
      event: event,
      body: encode(data, {useBigInt64: true})
    };
    console.log(`[MakeMsgData]`, buffer);
    const payload = encode(buffer, {useBigInt64: true});
    return new NetMsgData(msgId, Buffer.from(payload));
  }

}
