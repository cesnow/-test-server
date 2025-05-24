import Long from "long";
import moment from "moment";

import {MessageType, PartialMessage, reflectionMergePartial} from "@protobuf-ts/runtime";
import {getTProtoChecksum, newTObject} from "@/components/TProto/checksum";
import {Any} from "@/components/TProto/pb/any";

export class NetMsgData {
  msgId: Long;
  msgData: Buffer;

  constructor(msgId: Long, msgData: Buffer) {
    this.msgId = msgId;
    this.msgData = msgData;
  }
}

export class IncomeMessage {
  checksum: number;
  raw: MessageType<any>;
  data: any;

  constructor(checksum: number, raw: MessageType<any>, data: any) {
    this.checksum = checksum;
    this.raw = raw;
    this.data = data;
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

  public MakeMsgData<T extends object>(msgType: MessageType<T>, data: PartialMessage<T>): NetMsgData {
    const classId: number = getTProtoChecksum(msgType);

    const msg = msgType.create()
    reflectionMergePartial(msgType, msg, data);
    const body = msgType.toBinary(msg);

    let msgId: Long = this.NextMessageId();
    let seq: number = this.getSeq();
    let buf: Buffer = Buffer.alloc(8 + 8 + 8 + 4 + 4 + 4);
    buf.writeInt32LE(NetworkData.getInstance().AuthKeyId.low, 0);
    buf.writeInt32LE(NetworkData.getInstance().AuthKeyId.high, 4);
    buf.writeInt32LE(NetworkData.getInstance().SessionId.low, 8);
    buf.writeInt32LE(NetworkData.getInstance().SessionId.high, 12);
    buf.writeInt32LE(msgId.low, 16);
    buf.writeInt32LE(msgId.high, 20);
    buf.writeInt32LE(seq, 24);
    buf.writeInt32LE(body.length + 4, 28);
    buf.writeInt32LE(classId, 32);
    buf = Buffer.concat([buf, body]);

    console.log(`MsgId: ${msgId.toString()}`);
    console.log(`MakeMsgData: Header /// Data \n\t${buf.subarray(0, 8 + 8 + 8 + 8 + 4 + 4).toString('hex')} /// ${buf.subarray(8 + 8 + 8 + 8 + 4 + 4).toString('hex')}`);
    return new NetMsgData(msgId, buf);
  }

  public MakePackets(data: Buffer): Buffer {
    let sizeBuf: Buffer = Buffer.alloc(4);
    let size: number = data.length;
    sizeBuf.writeUInt32LE(size, 0);
    return Buffer.concat([sizeBuf, data]);
  }

  public ParseIncomeMessage(b: Buffer): IncomeMessage | null {
    console.log(`ParseFromIncomingMessage: ${b.toString('hex')}`);

    const checksum = b.readInt32LE(0);
    const msgType = newTObject(checksum);

    if (!msgType) {
      return null;
    }

    const payload = b.subarray(4);
    return new IncomeMessage(checksum, msgType, msgType.fromBinary(payload));
  }

  public ParseAnyToIncomeMessage(b: Any): IncomeMessage | null {
    console.log(b.typeUrl);

    return null;
  }
}










