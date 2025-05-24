import { encode } from "@msgpack/msgpack";
import Long from "long";
import moment from "moment";

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

  private lastOutgoingMessageId: Long = Long.ZERO;
  private sequenceNum: Long = Long.ZERO;
  public AuthKeyId: Long = Long.ZERO;
  public SessionId: Long = Long.ZERO;

  private constructor() {
  }

  public reset(): void {
    this.lastOutgoingMessageId = Long.ZERO;
    this.sequenceNum = Long.ZERO;
    this.AuthKeyId = Long.ZERO;
    this.SessionId = Long.ZERO;
  }

  public static getInstance(): NetworkData {
    if (!NetworkData.instance) {
      NetworkData.instance = new NetworkData();
    }
    return NetworkData.instance;
  }

  private getSeq(): number {
    this.sequenceNum.add(1);
    return this.sequenceNum.toNumber();
  }

  public NextMessageId(): Long {
    const nano: Long = Long.fromNumber(1000 * 1000 * 1000);
    const unixNano: Long = Long.fromNumber(moment().valueOf() * 1000000);
    const messageIdPart1: Long = unixNano.div(nano).shiftLeft(32);
    const messageIdPart2: Long = unixNano.mod(nano).and(Long.fromNumber(-4));
    let messageId: Long = messageIdPart1.or(messageIdPart2);
    // let messageId:Long = ((Math.floor(unixNano / nano) << 32) | (unixNano % nano)) & -4;
    if (messageId.lessThanOrEqual(this.lastOutgoingMessageId)) {
      messageId = this.lastOutgoingMessageId.add(1);
    }
    while (messageId.modulo(4).notEquals(0)) {
      messageId.add(1);
    }
    this.lastOutgoingMessageId = messageId;
    return messageId;
  }

  public MakeMsgData(event: string, data: object): NetMsgData {

    let msgId: Long = this.NextMessageId();
    let seq: number = this.getSeq();

    const buffer = {
      msgId: msgId.toBigInt(),
      seqNo: seq,
      event: event,
      body: encode(data, {useBigInt64: true})
    };
    const payload = encode(buffer, {useBigInt64: true});

    return new NetMsgData(msgId, Buffer.from(payload));
  }

}










